// auditlock.go -- cross-process serialization for audit log appends.
//
// Several agent harnesses (Claude Code, Codex, Kimi) run this guard
// concurrently on one machine. The append path is a read-modify-write
// cycle -- read the chain tail hash, then write the next line -- and
// without synchronization two processes can chain from the same tail, or
// both write a genesis line into a fresh log. Either way the hash chain
// that rr-audit-chain verifies reports BROKEN.
//
// The whole critical section (tail read, optional rotation, append) runs
// under an exclusive advisory flock(2) on a sidecar lock file at
// <log>.lock. A sidecar -- not the log itself -- because verify-guard
// rotation renames the log; the lock must pin a stable inode. The lock
// file is created once and never deleted: deleting it would let a new
// opener lock a fresh inode while a current holder still holds the old
// one. flock state dies with the fd, so a killed guard never leaves the
// lock held.
//
// Failure policy matches the rest of the audit path: if the lock cannot
// be taken, the append is skipped (the error is returned for strict
// callers) rather than performed unlocked. A lost audit line is
// recoverable; a broken chain is not. Policy enforcement never blocks on
// audit I/O.
//
// syscall.Flock is stdlib and available on darwin and linux; this fleet
// is unix-only.
package main

import (
	"os"
	"syscall"
)

// withAuditLock runs fn while holding an exclusive advisory lock on
// path+".lock", serializing competing guard processes. fn must contain
// the entire read-tail-hash -> append critical section (including any
// rotation), because flock is released as soon as fn returns.
func withAuditLock(path string, fn func() error) error {
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	return fn()
}
