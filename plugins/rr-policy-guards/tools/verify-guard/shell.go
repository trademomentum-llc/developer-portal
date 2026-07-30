package main

import "strings"

type targetCommandSet struct {
	Commit                 bool
	Push                   bool
	RepoDirs               []string
	RepoContextUnsupported bool
}

func (targets targetCommandSet) Any() bool {
	return targets.Commit || targets.Push
}

func detectTargetCommands(command string) targetCommandSet {
	var targets targetCommandSet
	scanShellCommands(command, &targets)
	return targets
}

func scanShellCommands(command string, targets *targetCommandSet) {
	var segment strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		classifyShellSegment(segment.String(), targets)
		segment.Reset()
	}

	for i := 0; i < len(command); i++ {
		char := command[i]
		if escaped {
			segment.WriteByte(char)
			escaped = false
			continue
		}
		if char == '\\' && !inSingle {
			segment.WriteByte(char)
			escaped = true
			continue
		}
		if char == '\'' && !inDouble {
			inSingle = !inSingle
			segment.WriteByte(char)
			continue
		}
		if char == '"' && !inSingle {
			inDouble = !inDouble
			segment.WriteByte(char)
			continue
		}
		if char == '$' && !inSingle && i+1 < len(command) && command[i+1] == '(' {
			if end, ok := commandSubstitutionEnd(command, i+2); ok {
				scanShellCommands(command[i+2:end], targets)
				segment.WriteString(" substitution ")
				i = end
				continue
			}
		}
		if char == '`' && !inSingle {
			if end, ok := backtickEnd(command, i+1); ok {
				scanShellCommands(command[i+1:end], targets)
				segment.WriteString(" substitution ")
				i = end
				continue
			}
		}
		if !inSingle && !inDouble && isShellSeparator(char) {
			flush()
			continue
		}
		segment.WriteByte(char)
	}
	flush()
}

func commandSubstitutionEnd(command string, start int) (int, bool) {
	depth := 1
	inSingle := false
	inDouble := false
	escaped := false
	for i := start; i < len(command); i++ {
		char := command[i]
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' && !inSingle {
			escaped = true
			continue
		}
		if char == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if char == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}
		if inSingle || inDouble {
			continue
		}
		switch char {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func backtickEnd(command string, start int) (int, bool) {
	escaped := false
	for i := start; i < len(command); i++ {
		char := command[i]
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' {
			escaped = true
			continue
		}
		if char == '`' {
			return i, true
		}
	}
	return 0, false
}

func isShellSeparator(char byte) bool {
	switch char {
	case ';', '&', '|', '\n', '(', ')':
		return true
	default:
		return false
	}
}

func classifyShellSegment(segment string, targets *targetCommandSet) {
	tokens := tokenizeShell(segment)
	index := commandTokenIndex(tokens)
	if index < 0 || index >= len(tokens) {
		return
	}
	if isDirectoryChangingCommand(tokens[index]) || wrapperChangesDirectory(tokens[:index]) {
		targets.RepoContextUnsupported = true
		return
	}
	if !isGitExecutable(tokens[index]) {
		return
	}
	subcommand, repoDir, unsupported := gitTarget(tokens[index+1:])
	if subcommand == "commit" {
		targets.Commit = true
	} else if subcommand == "push" {
		targets.Push = true
	} else {
		return
	}
	targets.RepoDirs = append(targets.RepoDirs, repoDir)
	if unsupported {
		targets.RepoContextUnsupported = true
	}
}

func commandTokenIndex(tokens []string) int {
	index := 0
	for index < len(tokens) && isShellAssignment(tokens[index]) {
		index++
	}
	for index < len(tokens) {
		switch tokens[index] {
		case "command", "builtin", "exec":
			index++
		case "env":
			index++
			for index < len(tokens) {
				token := tokens[index]
				if isShellAssignment(token) {
					index++
					continue
				}
				switch token {
				case "-u", "--unset", "-C", "--chdir", "-S", "--split-string":
					index += 2
				case "-i", "--ignore-environment", "-0", "--null":
					index++
				default:
					if strings.HasPrefix(token, "--unset=") || strings.HasPrefix(token, "--chdir=") || strings.HasPrefix(token, "--split-string=") {
						index++
						continue
					}
					return index
				}
			}
		case "sudo":
			index++
			for index < len(tokens) {
				token := tokens[index]
				if !strings.HasPrefix(token, "-") {
					return index
				}
				switch token {
				case "-D", "--chdir", "-u", "--user", "-g", "--group", "-h", "--host", "-p", "--prompt", "-r", "--role", "-t", "--type", "-C", "--close-from":
					index += 2
				default:
					index++
				}
			}
		default:
			return index
		}
	}
	return -1
}

func isShellAssignment(token string) bool {
	equals := strings.IndexByte(token, '=')
	if equals <= 0 {
		return false
	}
	for index := 0; index < equals; index++ {
		char := token[index]
		if index == 0 {
			if !isShellIdentStart(char) {
				return false
			}
		} else if !isShellIdentChar(char) {
			return false
		}
	}
	return true
}

func isShellIdentStart(char byte) bool {
	return char == '_' || char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z'
}

func isShellIdentChar(char byte) bool {
	return isShellIdentStart(char) || char >= '0' && char <= '9'
}

func isGitExecutable(token string) bool {
	return token == "git" || strings.HasSuffix(token, "/git")
}

func isDirectoryChangingCommand(token string) bool {
	return token == "cd" || token == "pushd" || token == "popd"
}

func wrapperChangesDirectory(tokens []string) bool {
	for index, token := range tokens {
		if token == "--chdir" || strings.HasPrefix(token, "--chdir=") || token == "-D" {
			return true
		}
		if token == "-C" && index > 0 && tokens[index-1] == "env" {
			return true
		}
	}
	return false
}

func gitTarget(arguments []string) (subcommand, repoDir string, unsupported bool) {
	seenRepoDir := false
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		switch argument {
		case "-C":
			if index+1 >= len(arguments) {
				return "", "", true
			}
			if seenRepoDir {
				unsupported = true
			}
			repoDir = arguments[index+1]
			seenRepoDir = true
			index++
			continue
		case "-c", "--config-env":
			index++
			continue
		case "--git-dir", "--work-tree", "--namespace", "--super-prefix":
			unsupported = true
			index++
			continue
		}
		if strings.HasPrefix(argument, "--git-dir=") || strings.HasPrefix(argument, "--work-tree=") || strings.HasPrefix(argument, "--namespace=") || strings.HasPrefix(argument, "--super-prefix=") {
			unsupported = true
			continue
		}
		if strings.HasPrefix(argument, "-") {
			continue
		}
		return argument, repoDir, unsupported
	}
	return "", repoDir, unsupported
}

func tokenizeShell(command string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	for index := 0; index < len(command); index++ {
		char := command[index]
		switch {
		case inSingle:
			if char == '\'' {
				inSingle = false
			} else {
				current.WriteByte(char)
			}
		case inDouble:
			if char == '"' {
				inDouble = false
			} else if char == '\\' && index+1 < len(command) {
				next := command[index+1]
				if next == '"' || next == '\\' || next == '$' || next == '`' {
					current.WriteByte(next)
					index++
				} else {
					current.WriteByte(char)
				}
			} else {
				current.WriteByte(char)
			}
		default:
			switch char {
			case '\'':
				inSingle = true
			case '"':
				inDouble = true
			case '\\':
				if index+1 < len(command) {
					current.WriteByte(command[index+1])
					index++
				}
			case ' ', '\t':
				flush()
			default:
				current.WriteByte(char)
			}
		}
	}
	flush()
	return tokens
}
