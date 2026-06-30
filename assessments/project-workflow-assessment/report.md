# Project Workflow Assessment

Workspace root: `/Users/nexus1/Projects`

## Summary

- Projects analyzed: 169
- Status mix: healthy=115, watch=42, critical=12
- Projects with low-fit datasets: 4
- Projects with cross-project duplicate datasets: 5
- Projects with checked-in artifacts: 39
- Projects with private keys or literal credential hits: 50

## Workflow Glossary

- `THE PROJECT RECON BLITZ`: Inventory layout, manifests, languages, docs, tests, and data assets.
- `THE BUILD TRUTH SERUM`: Surface missing or weak build, test, lint, and type-check coverage.
- `THE SECRET HUNTER`: Scan for private keys, literal credentials, PII-bearing data, and hardcoded paths.
- `THE ARTIFACT PURGE`: Flag checked-in runtime and build artifacts before they poison review and search.
- `THE DATA PREFLIGHT`: Classify project datasets before handing them to an LLM or analytics workflow.
- `THE DUPLICATE TREE HUNTER`: Catch mirrored files, duplicated splits, and copied data across sibling projects.

## Primary Workflow Count

- `THE ARTIFACT PURGE`: 9
- `THE BUILD TRUTH SERUM`: 90
- `THE DATA PREFLIGHT`: 1
- `THE PROJECT RECON BLITZ`: 19
- `THE SECRET HUNTER`: 50

## Project Inventory

| Project | Kind | Score | Code | Tests | Datasets | Artifacts | Primary Workflow | Notes |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| agents/OpenHands-CLI | python_project | 80 | 250 | 218 | 0/0 | 0 | THE SECRET HUNTER | Literal credential-like assignments were detected in text files. |
| agents/agentkit | docs_or_code_project | 95 | 23 | 3 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/python/coinbase-agentkit | python_project | 50 | 292 | 161 | 0/0 | 0 | THE SECRET HUNTER | Private key material appears inside the project tree. Literal credential-like assignments were detected in text files. |
| agents/agentkit/python/coinbase-agentkit/docs | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/python/create-onchain-agent | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/advertisements-agent-with-strands-agents-x402-cdp-chatbot | python_project | 100 | 2 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/autogen-cdp-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/langchain-cdp-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/langchain-cdp-chatbot-with-amazon-bedrock | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/langchain-cdp-smart-wallet-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/langchain-cdp-solana-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/langchain-eth-account-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/langchain-nillion-secretvault-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/langchain-twitter-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/openai-agents-sdk-cdp-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/openai-agents-sdk-cdp-voice-chatbot | python_project | 100 | 2 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/openai-agents-sdk-smart-wallet-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/pydantic-ai-cdp-chatbot | python_project | 100 | 3 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/examples/strands-agents-cdp-server-chatbot | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/framework-extensions/autogen | python_project | 100 | 6 | 4 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| agents/agentkit/python/framework-extensions/langchain | python_project | 100 | 2 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/framework-extensions/langchain/docs | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/python/framework-extensions/openai-agents-sdk | python_project | 100 | 6 | 4 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| agents/agentkit/python/framework-extensions/pydantic-ai | python_project | 100 | 6 | 4 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| agents/agentkit/python/framework-extensions/strands-agents | python_project | 100 | 2 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| agents/agentkit/python/framework-extensions/strands-agents/docs | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript | node_app | 100 | 0 | 0 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| agents/agentkit/typescript/agentkit | node_app | 50 | 269 | 59 | 0/0 | 0 | THE SECRET HUNTER | Private key material appears inside the project tree. Literal credential-like assignments were detected in text files. |
| agents/agentkit/typescript/create-onchain-agent | node_app | 85 | 21 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/create-onchain-agent/templates/mcp | node_app | 100 | 11 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/create-onchain-agent/templates/next | node_app | 85 | 20 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/langchain-cdp-chatbot | node_app | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/langchain-cdp-smart-wallet-chatbot | node_app | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/langchain-farcaster-chatbot | node_app | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/langchain-privy-chatbot | node_app | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/langchain-solana-chatbot | node_app | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/langchain-twitter-chatbot | node_app | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/langchain-xmtp-chatbot | node_app | 100 | 2 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/langchain-zerodev-chatbot | node_app | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/model-context-protocol-smart-wallet-server | node_app | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/examples/vercel-ai-sdk-smart-wallet-chatbot | node_app | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/agentkit/typescript/framework-extensions/langchain | node_app | 100 | 3 | 1 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| agents/agentkit/typescript/framework-extensions/model-context-protocol | node_app | 100 | 3 | 1 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| agents/agentkit/typescript/framework-extensions/vercel-ai-sdk | node_app | 100 | 4 | 1 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| agents/claude-plugins-official | docs_or_code_project | 80 | 36 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| agents/software-agent-sdk | python_project | 80 | 534 | 452 | 0/0 | 0 | THE SECRET HUNTER | Literal credential-like assignments were detected in text files. |
| agents/software-agent-sdk/openhands-agent-server | python_project | 65 | 35 | 0 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. Literal credential-like assignments were detected in text files. |
| agents/software-agent-sdk/openhands-agent-server/openhands/agent_server/vscode_extensions/openhands-settings | node_app | 90 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| agents/software-agent-sdk/openhands-sdk | python_project | 80 | 203 | 1 | 0/0 | 0 | THE SECRET HUNTER | Literal credential-like assignments were detected in text files. |
| agents/software-agent-sdk/openhands-tools | python_project | 65 | 85 | 0 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. Literal credential-like assignments were detected in text files. |
| agents/software-agent-sdk/openhands-workspace | python_project | 80 | 10 | 0 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. Literal credential-like assignments were detected in text files. |
| engine | compiled_project | 77 | 21 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. Hardcoded workspace paths found: 1. |
| frameworks | python_project | 32 | 65 | 9 | 0/0 | 17 | THE SECRET HUNTER | Checked-in artifact dirs: 17. Hardcoded workspace paths found: 12. Private key material appears inside the project tree. |
| infra/Conscious Living System | python_project | 90 | 18 | 1 | 0/0 | 10 | THE ARTIFACT PURGE | Checked-in artifact dirs: 10. |
| infra/frameworks | docs_or_code_project | 77 | 14 | 0 | 0/0 | 3 | THE ARTIFACT PURGE | Checked-in artifact dirs: 3. Code is present but test coverage surface is thin or not wired into inferred commands. Hardcoded workspace paths found: 4. |
| omni/agentxfoundry | docs_or_code_project | 65 | 38 | 1 | 0/0 | 11 | THE SECRET HUNTER | Checked-in artifact dirs: 11. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/agentxfoundry/legacy | node_app | 70 | 3 | 0 | 0/0 | 1 | THE SECRET HUNTER | Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/arch-alive/qenv/lib/python3.14/site-packages/mistralai/extra | docs_or_code_project | 65 | 21 | 3 | 0/0 | 6 | THE SECRET HUNTER | Checked-in artifact dirs: 6. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/cali/jeff/deepfiat | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/deepfiat/backend | node_app | 80 | 20 | 1 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/cali/jeff/deepfiat/backend/c | compiled_project | 100 | 19 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/deepfiat/frontend | node_app | 55 | 24 | 0 | 0/0 | 1 | THE SECRET HUNTER | Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/cali/jeff/deepfiat/zk | node_app | 100 | 4 | 1 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/cali/jeff/theratouch+ | node_app | 62 | 18 | 160 | 0/0 | 1 | THE SECRET HUNTER | Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. Hardcoded workspace paths found: 20. |
| omni/cali/jeff/theratouch+/vendor/laravel/framework/src/Illuminate/Foundation/resources/exceptions/renderer | node_app | 90 | 2 | 0 | 0/0 | 1 | THE ARTIFACT PURGE | Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/theratouch+/vendor/mockery/mockery/docs | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar | docs_or_code_project | 87 | 1 | 0 | 1/0 | 0 | THE DATA PREFLIGHT | Datasets: 1 unique groups, 0 low-fit for raw LLM analysis. Code is present but test coverage surface is thin or not wired into inferred commands. Hardcoded workspace paths found: 24. |
| omni/cali/jeff/tri-mar/platform | compiled_project | 40 | 78 | 1 | 0/0 | 16 | THE SECRET HUNTER | Checked-in artifact dirs: 16. Private key material appears inside the project tree. Literal credential-like assignments were detected in text files. |
| omni/cali/jeff/tri-mar/platform/backend | node_app | 70 | 101 | 13 | 0/0 | 2 | THE SECRET HUNTER | Checked-in artifact dirs: 2. Literal credential-like assignments were detected in text files. |
| omni/cali/jeff/tri-mar/platform/cmd/fiat-btc | generic_build_project | 100 | 5 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/forge-std | node_app | 90 | 1 | 23 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts | node_app | 80 | 269 | 228 | 0/0 | 0 | THE SECRET HUNTER | Literal credential-like assignments were detected in text files. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable | node_app | 80 | 269 | 228 | 0/0 | 0 | THE SECRET HUNTER | Literal credential-like assignments were detected in text files. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/contracts | node_app | 100 | 0 | 0 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/fv | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/forge-std | node_app | 90 | 1 | 23 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts | node_app | 80 | 269 | 228 | 0/0 | 0 | THE SECRET HUNTER | Literal credential-like assignments were detected in text files. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/contracts | node_app | 100 | 0 | 0 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/fv | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/lib/forge-std | node_app | 90 | 1 | 23 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/scripts/solhint-custom | node_app | 90 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/scripts/solhint-custom | node_app | 90 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts/contracts | node_app | 100 | 0 | 0 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts/fv | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts/lib/forge-std | node_app | 90 | 1 | 23 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts/scripts/solhint-custom | node_app | 90 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/cali/jeff/tri-mar/platform/frontend | node_app | 75 | 48 | 0 | 0/0 | 2 | THE ARTIFACT PURGE | Checked-in artifact dirs: 2. Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/vps/network | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/vps/network/ebpf | generic_build_project | 100 | 3 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/vps/network/tor | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/vps/orchestrator | generic_build_project | 100 | 6 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/vps/proto | generic_build_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/cali/jeff/tri-mar/platform/zk | node_app | 90 | 3 | 0 | 0/0 | 1 | THE ARTIFACT PURGE | Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/crdroid-source/.repo/repo | python_project | 70 | 82 | 26 | 0/0 | 2 | THE SECRET HUNTER | Checked-in artifact dirs: 2. Literal credential-like assignments were detected in text files. |
| omni/distributed-ai-cluster | python_project | 32 | 148 | 5 | 1/0 | 21 | THE SECRET HUNTER | Checked-in artifact dirs: 21. Datasets: 1 unique groups, 0 low-fit for raw LLM analysis. Contains contact-style data fields; redact before external model use. |
| omni/distributed-ai-cluster/dashboard | node_app | 70 | 138 | 1 | 0/0 | 2 | THE SECRET HUNTER | Checked-in artifact dirs: 2. Literal credential-like assignments were detected in text files. |
| omni/distributed-ai-cluster/legacy | node_app | 70 | 3 | 0 | 0/0 | 1 | THE SECRET HUNTER | Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/eight-layer-pqc | python_project | 60 | 16 | 3 | 5/0 | 2 | THE SECRET HUNTER | Checked-in artifact dirs: 2. Datasets: 5 unique groups, 0 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 5. |
| omni/eight-layer-pqc/code/go | generic_build_project | 100 | 3 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/eight-layer-pqc/code/rust | generic_build_project | 100 | 3 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights | docs_or_code_project | 5 | 214 | 5 | 22/11 | 5 | THE SECRET HUNTER | Checked-in artifact dirs: 5. Datasets: 22 unique groups, 11 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 21. |
| omni/health-insights/app_backup | node_app | 65 | 24 | 0 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/health-insights/applications | node_app | 70 | 35 | 1 | 0/0 | 4 | THE SECRET HUNTER | Checked-in artifact dirs: 4. Literal credential-like assignments were detected in text files. |
| omni/health-insights/applications/frontend | node_app | 70 | 9 | 0 | 0/0 | 2 | THE SECRET HUNTER | Checked-in artifact dirs: 2. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/health-insights/ihep-application | node_app | 40 | 351 | 13 | 21/11 | 3 | THE SECRET HUNTER | Checked-in artifact dirs: 3. Datasets: 21 unique groups, 11 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 21. |
| omni/health-insights/ihep-application/hub/gateway | node_app | 70 | 19 | 0 | 0/0 | 4 | THE SECRET HUNTER | Checked-in artifact dirs: 4. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/health-insights/terraform/.terraform/modules/autokey-org-policy-gcp-restrict-cmek-crypto-key-projects | generic_build_project | 100 | 0 | 23 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/health-insights/terraform/.terraform/modules/autokey-org-policy-gcp-restrict-cmek-crypto-key-projects/test/integration | generic_build_project | 100 | 2 | 12 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/terraform/.terraform/modules/cs-autokey | generic_build_project | 100 | 4 | 11 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/terraform/.terraform/modules/cs-autokey/test/integration | generic_build_project | 100 | 4 | 6 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/terraform/.terraform/modules/cs-common | generic_build_project | 100 | 1 | 12 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/terraform/.terraform/modules/cs-folders-iam-0-computeinstanceAdminv1 | generic_build_project | 100 | 0 | 44 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/health-insights/terraform/.terraform/modules/cs-folders-iam-0-computeinstanceAdminv1/test/integration | generic_build_project | 100 | 0 | 16 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/health-insights/terraform/.terraform/modules/cs-gg-lifescien-concept-nonprod-svc | generic_build_project | 100 | 0 | 11 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/health-insights/terraform/.terraform/modules/cs-kms-projects | generic_build_project | 100 | 9 | 41 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/terraform/.terraform/modules/cs-kms-projects/test/integration | generic_build_project | 100 | 4 | 30 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/terraform/.terraform/modules/cs-logging-destination | generic_build_project | 100 | 9 | 49 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/terraform/.terraform/modules/cs-logging-destination/modules/bq-log-alerting/logging/cloud_function | node_app | 90 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/health-insights/terraform/.terraform/modules/cs-logging-destination/test/integration | generic_build_project | 100 | 4 | 36 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/terraform/.terraform/modules/cs-vpc-dev-shared | generic_build_project | 100 | 1 | 32 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/terraform/.terraform/modules/cs-vpc-dev-shared/test/integration | generic_build_project | 100 | 15 | 17 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/health-insights/workspaces | node_app | 80 | 15 | 0 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/health-insights/workspaces/frontend | node_app | 80 | 8 | 0 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/ihep-app | node_app | 45 | 5113 | 0 | 0/0 | 6 | THE SECRET HUNTER | Checked-in artifact dirs: 6. Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/ihep-app/ihep | node_app | 40 | 221 | 10 | 21/11 | 5 | THE SECRET HUNTER | Checked-in artifact dirs: 5. Datasets: 21 unique groups, 11 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 21. |
| omni/ihep-app/ihep/app | node_app | 80 | 15 | 0 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/ihep-app/ihep/app/frontend | node_app | 80 | 8 | 0 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/ihep-app/ihep/applications | node_app | 70 | 35 | 1 | 0/0 | 6 | THE SECRET HUNTER | Checked-in artifact dirs: 6. Literal credential-like assignments were detected in text files. |
| omni/ihep-app/ihep/applications/frontend | node_app | 70 | 9 | 0 | 0/0 | 2 | THE SECRET HUNTER | Checked-in artifact dirs: 2. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| omni/ihep-app/ihep/terraform/.terraform/modules/autokey-org-policy-gcp-restrict-cmek-crypto-key-projects | generic_build_project | 100 | 0 | 23 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/ihep-app/ihep/terraform/.terraform/modules/autokey-org-policy-gcp-restrict-cmek-crypto-key-projects/test/integration | generic_build_project | 100 | 2 | 12 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-autokey | generic_build_project | 100 | 4 | 11 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-autokey/test/integration | generic_build_project | 100 | 4 | 6 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-common | generic_build_project | 100 | 1 | 12 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-folders-iam-0-computeinstanceAdminv1 | generic_build_project | 100 | 0 | 44 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-folders-iam-0-computeinstanceAdminv1/test/integration | generic_build_project | 100 | 0 | 16 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-gg-lifescien-concept-nonprod-svc | generic_build_project | 100 | 0 | 11 | 0/0 | 0 | THE PROJECT RECON BLITZ | - |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-kms-projects | generic_build_project | 100 | 9 | 41 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-kms-projects/test/integration | generic_build_project | 100 | 4 | 30 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-logging-destination | generic_build_project | 100 | 9 | 49 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-logging-destination/modules/bq-log-alerting/logging/cloud_function | node_app | 90 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-logging-destination/test/integration | generic_build_project | 100 | 4 | 36 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-vpc-dev-shared | generic_build_project | 100 | 1 | 32 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/ihep-app/ihep/terraform/.terraform/modules/cs-vpc-dev-shared/test/integration | generic_build_project | 100 | 15 | 17 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/jarmacz.com | node_app | 55 | 42 | 0 | 1/0 | 3 | THE SECRET HUNTER | Checked-in artifact dirs: 3. Datasets: 1 unique groups, 0 low-fit for raw LLM analysis. Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/multimodal-agent-builder | mixed_node_python | 62 | 185 | 12 | 17/2 | 8 | THE SECRET HUNTER | Checked-in artifact dirs: 8. Datasets: 17 unique groups, 2 low-fit for raw LLM analysis. Literal credential-like assignments were detected in text files. |
| omni/neurodivergence.works/security-architecture/eight-layer-pqc | docs_or_code_project | 65 | 14 | 3 | 5/0 | 0 | THE SECRET HUNTER | Datasets: 5 unique groups, 0 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 5. Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/neurodivergence.works/security-architecture/eight-layer-pqc/code/go | generic_build_project | 100 | 3 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/neurodivergence.works/security-architecture/eight-layer-pqc/code/rust | generic_build_project | 100 | 3 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| omni/provenance-engine | python_project | 90 | 18 | 3 | 0/0 | 3 | THE ARTIFACT PURGE | Checked-in artifact dirs: 3. |
| omni/provenance-engine/build/pkg/payload/opt/provenance-engine | python_project | 100 | 3 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| omni/security-workspace | docs_or_code_project | 65 | 8 | 0 | 1/0 | 1 | THE SECRET HUNTER | Checked-in artifact dirs: 1. Datasets: 1 unique groups, 0 low-fit for raw LLM analysis. Code is present but test coverage surface is thin or not wired into inferred commands. |
| tools/foundational-frameworks | node_app | 90 | 32 | 2 | 0/0 | 2 | THE ARTIFACT PURGE | Checked-in artifact dirs: 2. |
| tools/gemini-cli-0.22.2 | node_app | 70 | 73 | 32 | 0/0 | 1 | THE SECRET HUNTER | Checked-in artifact dirs: 1. Literal credential-like assignments were detected in text files. |
| tools/gemini-cli-0.22.2/packages/a2a-server | node_app | 90 | 30 | 9 | 0/0 | 2 | THE ARTIFACT PURGE | Checked-in artifact dirs: 2. |
| tools/gemini-cli-0.22.2/packages/cli | node_app | 70 | 649 | 339 | 0/0 | 1 | THE SECRET HUNTER | Checked-in artifact dirs: 1. Literal credential-like assignments were detected in text files. |
| tools/gemini-cli-0.22.2/packages/cli/src/commands/extensions/examples/mcp-server | node_app | 100 | 2 | 1 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| tools/gemini-cli-0.22.2/packages/core | node_app | 40 | 449 | 199 | 0/0 | 1 | THE SECRET HUNTER | Checked-in artifact dirs: 1. Private key material appears inside the project tree. Literal credential-like assignments were detected in text files. |
| tools/gemini-cli-0.22.2/packages/test-utils | node_app | 100 | 4 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| tools/gemini-cli-0.22.2/packages/vscode-ide-companion | node_app | 90 | 11 | 3 | 0/0 | 1 | THE ARTIFACT PURGE | Checked-in artifact dirs: 1. |
| tools/gemini-cli-0.22.2/third_party/get-ripgrep | node_app | 100 | 2 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. |
| tools/github-mcp-server | generic_build_project | 80 | 58 | 1 | 0/0 | 0 | THE SECRET HUNTER | Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |
| tools/mcp-server | python_project | 80 | 63 | 51 | 0/0 | 0 | THE SECRET HUNTER | Literal credential-like assignments were detected in text files. |
| tools/mcpb | node_app | 70 | 39 | 11 | 0/0 | 2 | THE SECRET HUNTER | Checked-in artifact dirs: 2. Literal credential-like assignments were detected in text files. |
| tools/mcpb/examples/file-manager-python | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| tools/mcpb/examples/file-system-node | node_app | 90 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| tools/mcpb/examples/hello-world-node | node_app | 90 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. |
| tools/mcpb/examples/hello-world-uv | python_project | 100 | 1 | 0 | 0/0 | 0 | THE BUILD TRUTH SERUM | Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. |
| tools/skills | docs_or_code_project | 65 | 49 | 1 | 0/0 | 2 | THE SECRET HUNTER | Checked-in artifact dirs: 2. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files. |

## Detailed Findings

### `agents/OpenHands-CLI`

- Name: `openhands`
- Status: `watch` (score `80`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: AGENTS.md, README.md
- Languages: Python, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER
- Inferred commands: pytest, ruff check ., make build, make test, make clean
- Secret hits: api_key in agents/OpenHands-CLI/tests/test_agent_context_os_info.py, api_key in agents/OpenHands-CLI/tests/conftest.py, api_key in agents/OpenHands-CLI/tests/test_directory_separation.py, api_key in agents/OpenHands-CLI/tests/auth/test_api_client.py, api_key in agents/OpenHands-CLI/tests/auth/test_login_command.py
- Notes: Literal credential-like assignments were detected in text files.

### `agents/agentkit`

- Name: `agentkit`
- Status: `healthy` (score `95`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: README.md
- Languages: Python, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/python/coinbase-agentkit`

- Name: `coinbase-agentkit`
- Status: `critical` (score `50`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER
- Inferred commands: pytest, ruff check ., mypy ., make test
- Private key paths: agents/agentkit/python/coinbase-agentkit/tests/action_providers/ssh/test_ssh_connect.py
- Secret hits: api_key in agents/agentkit/python/coinbase-agentkit/tests/action_providers/twitter/test_action_provider.py, password in agents/agentkit/python/coinbase-agentkit/tests/action_providers/ssh/conftest.py, password in agents/agentkit/python/coinbase-agentkit/tests/action_providers/ssh/test_connection.py, password in agents/agentkit/python/coinbase-agentkit/tests/action_providers/ssh/test_connection_pool.py, password in agents/agentkit/python/coinbase-agentkit/tests/action_providers/ssh/test_keys.py
- Notes: Private key material appears inside the project tree. Literal credential-like assignments were detected in text files.

### `agents/agentkit/python/coinbase-agentkit/docs`

- Name: `docs`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: none
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/python/create-onchain-agent`

- Name: `create-onchain-agent`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/advertisements-agent-with-strands-agents-x402-cdp-chatbot`

- Name: `ads-agent`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/autogen-cdp-chatbot`

- Name: `chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/langchain-cdp-chatbot`

- Name: `chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/langchain-cdp-chatbot-with-amazon-bedrock`

- Name: `chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/langchain-cdp-smart-wallet-chatbot`

- Name: `chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/langchain-cdp-solana-chatbot`

- Name: `langchain-cdp-solana-chatbot`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/langchain-eth-account-chatbot`

- Name: `chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/langchain-nillion-secretvault-chatbot`

- Name: `chatbot-python-with-nillion-secretvault`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/langchain-twitter-chatbot`

- Name: `chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/openai-agents-sdk-cdp-chatbot`

- Name: `openai-agents-sdk-chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/openai-agents-sdk-cdp-voice-chatbot`

- Name: `openai-agents-cdp-voice-chatbot`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/openai-agents-sdk-smart-wallet-chatbot`

- Name: `openai-agents-sdk-chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/pydantic-ai-cdp-chatbot`

- Name: `pydantic-ai-chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/examples/strands-agents-cdp-server-chatbot`

- Name: `chatbot-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check .
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/framework-extensions/autogen`

- Name: `coinbase-agentkit-autogen`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: pytest, ruff check ., mypy ., make test

### `agents/agentkit/python/framework-extensions/langchain`

- Name: `coinbase-agentkit-langchain`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check ., mypy ., make test
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/framework-extensions/langchain/docs`

- Name: `docs`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: none
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/python/framework-extensions/openai-agents-sdk`

- Name: `coinbase-agentkit-openai-agents-sdk`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: pytest, ruff check ., mypy ., make test

### `agents/agentkit/python/framework-extensions/pydantic-ai`

- Name: `coinbase-agentkit-pydantic-ai`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: pytest, ruff check ., mypy ., make test

### `agents/agentkit/python/framework-extensions/strands-agents`

- Name: `coinbase-agentkit-strands-agents`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: ruff check ., mypy ., make test
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `agents/agentkit/python/framework-extensions/strands-agents/docs`

- Name: `docs`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: none
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript`

- Name: `agentkit-typescript-monorepo`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: pnpm run dev, pnpm run build, pnpm run lint, pnpm run test

### `agents/agentkit/typescript/agentkit`

- Name: `@coinbase/agentkit`
- Status: `critical` (score `50`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test
- Private key paths: agents/agentkit/typescript/agentkit/src/wallet-providers/privyEvmDelegatedEmbeddedWalletProvider.ts
- Secret hits: api_key in agents/agentkit/typescript/agentkit/src/action-providers/farcaster/farcasterActionProvider.ts, api_key in agents/agentkit/typescript/agentkit/src/action-providers/farcaster/farcasterActionProvider.test.ts
- Notes: Private key material appears inside the project tree. Literal credential-like assignments were detected in text files.

### `agents/agentkit/typescript/create-onchain-agent`

- Name: `create-onchain-agent`
- Status: `healthy` (score `85`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/create-onchain-agent/templates/mcp`

- Name: `mcp`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run build
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/create-onchain-agent/templates/next`

- Name: `next-template`
- Status: `healthy` (score `85`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/langchain-cdp-chatbot`

- Name: `@coinbase/cdp-v2-evm-langchain-chatbot-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/langchain-cdp-smart-wallet-chatbot`

- Name: `@coinbase/smart-wallet-langchain-chatbot-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/langchain-farcaster-chatbot`

- Name: `@coinbase/farcaster-langchain-chatbot-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/langchain-privy-chatbot`

- Name: `@coinbase/langchain-privy-chatbot-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/langchain-solana-chatbot`

- Name: `@coinbase/solana-langchain-chatbot-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/langchain-twitter-chatbot`

- Name: `@coinbase/twitter-langchain-chatbot-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/langchain-xmtp-chatbot`

- Name: `@coinbase/langchain-xmtp-chatbot-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/langchain-zerodev-chatbot`

- Name: `@coinbase/langchain-zerodev-chatbot-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/model-context-protocol-smart-wallet-server`

- Name: `@coinbase/cdp-model-context-protocol-server-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run build, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/examples/vercel-ai-sdk-smart-wallet-chatbot`

- Name: `@coinbase/cdp-vercel-ai-sdk-chatbot-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run lint
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/agentkit/typescript/framework-extensions/langchain`

- Name: `@coinbase/agentkit-langchain`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test

### `agents/agentkit/typescript/framework-extensions/model-context-protocol`

- Name: `@coinbase/agentkit-model-context-protocol`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test

### `agents/agentkit/typescript/framework-extensions/vercel-ai-sdk`

- Name: `@coinbase/agentkit-vercel-ai-sdk`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test

### `agents/claude-plugins-official`

- Name: `claude-plugins-official`
- Status: `watch` (score `80`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: README.md
- Languages: Python, Shell, TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `agents/software-agent-sdk`

- Name: `software-agent-sdk`
- Status: `watch` (score `80`)
- Kind: `python_project`
- Manifests: Makefile, pyproject.toml
- Docs: AGENTS.md, README.md
- Languages: Python, Shell, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER
- Inferred commands: pytest, ruff check ., make build, make clean
- Secret hits: api_key in agents/software-agent-sdk/tests/conftest.py, api_key in agents/software-agent-sdk/tests/test_agent_step_bounded_scan.py, api_key in agents/software-agent-sdk/tests/tools/test_builtin_agents.py, api_key in agents/software-agent-sdk/tests/tools/test_working_dir_standardization.py, api_key in agents/software-agent-sdk/tests/tools/file_editor/test_file_editor_tool.py
- Notes: Literal credential-like assignments were detected in text files.

### `agents/software-agent-sdk/openhands-agent-server`

- Name: `openhands-agent-server`
- Status: `watch` (score `65`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: AGENTS.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Secret hits: api_key in agents/software-agent-sdk/openhands-agent-server/openhands/agent_server/conversation_router.py, api_key in agents/software-agent-sdk/openhands-agent-server/openhands/agent_server/conversation_router_acp.py
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. Literal credential-like assignments were detected in text files.

### `agents/software-agent-sdk/openhands-agent-server/openhands/agent_server/vscode_extensions/openhands-settings`

- Name: `openhands-settings`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `agents/software-agent-sdk/openhands-sdk`

- Name: `openhands-sdk`
- Status: `watch` (score `80`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: none
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER
- Inferred commands: pytest
- Secret hits: api_key in agents/software-agent-sdk/openhands-sdk/openhands/sdk/llm/llm.py, api_key in agents/software-agent-sdk/openhands-sdk/openhands/sdk/llm/auth/openai.py, api_key in agents/software-agent-sdk/openhands-sdk/openhands/sdk/llm/utils/litellm_provider.py, api_key in agents/software-agent-sdk/openhands-sdk/openhands/sdk/security/grayswan/analyzer.py, api_key in agents/software-agent-sdk/openhands-sdk/openhands/sdk/workspace/workspace.py
- Notes: Literal credential-like assignments were detected in text files.

### `agents/software-agent-sdk/openhands-tools`

- Name: `openhands-tools`
- Status: `watch` (score `65`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: none
- Languages: Python, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Secret hits: api_key in agents/software-agent-sdk/openhands-tools/openhands/tools/tom_consult/definition.py, api_key in agents/software-agent-sdk/openhands-tools/openhands/tools/tom_consult/executor.py
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. Literal credential-like assignments were detected in text files.

### `agents/software-agent-sdk/openhands-workspace`

- Name: `openhands-workspace`
- Status: `watch` (score `80`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: none
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Secret hits: api_key in agents/software-agent-sdk/openhands-workspace/openhands/workspace/remote_api/workspace.py, api_key in agents/software-agent-sdk/openhands-workspace/openhands/workspace/cloud/workspace.py
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree. Literal credential-like assignments were detected in text files.

### `engine`

- Name: `engine`
- Status: `watch` (score `77`)
- Kind: `compiled_project`
- Manifests: Makefile
- Docs: none
- Languages: C, C++
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: make, make clean
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. Hardcoded workspace paths found: 1.

### `frameworks`

- Name: `foundational-frameworks`
- Status: `critical` (score `32`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: pytest
- Artifact dirs: frameworks/.pytest_cache, frameworks/__pycache__, frameworks/code_check/__pycache__, frameworks/compliance/__pycache__, frameworks/core/__pycache__
- Private key paths: frameworks/.pqc-keys/0ffd8b8855d5b18f1baa1dae83c499cc.key, frameworks/.pqc-keys/c9e2c66609d33db6d0a46fab4638ca74.key, frameworks/.pqc-keys/e4e12033bf6b31a8e11ddfdd36eda46d.key, frameworks/.pqc-keys/a3f0c25ffb40aa20b263c99eabea7b3f.key, frameworks/.pqc-keys/a136c9ccb494dd149a82964a88b6a3e1.key
- Secret hits: secret in frameworks/pqc/hybrid_crypto.py
- Notes: Checked-in artifact dirs: 17. Hardcoded workspace paths found: 12. Private key material appears inside the project tree. Literal credential-like assignments were detected in text files.

### `infra/Conscious Living System`

- Name: `conscious-os`
- Status: `healthy` (score `90`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE ARTIFACT PURGE
- Inferred commands: pytest, ruff check ., black --check ., mypy .
- Artifact dirs: infra/Conscious Living System/.pytest_cache, infra/Conscious Living System/examples/__pycache__, infra/Conscious Living System/src/conscious_os/__pycache__, infra/Conscious Living System/src/conscious_os/core/__pycache__, infra/Conscious Living System/src/conscious_os/engines/__pycache__
- Notes: Checked-in artifact dirs: 10.

### `infra/frameworks`

- Name: `frameworks`
- Status: `watch` (score `77`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Artifact dirs: infra/frameworks/__pycache__, infra/frameworks/core/__pycache__, infra/frameworks/security/__pycache__
- Notes: Checked-in artifact dirs: 3. Code is present but test coverage surface is thin or not wired into inferred commands. Hardcoded workspace paths found: 4.

### `omni/agentxfoundry`

- Name: `agentxfoundry`
- Status: `watch` (score `65`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: CLAUDE.md, README.md
- Languages: Python, SQL
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Artifact dirs: omni/agentxfoundry/__pycache__, omni/agentxfoundry/agent_env, omni/agentxfoundry/agents/__pycache__, omni/agentxfoundry/agents/predictive/__pycache__, omni/agentxfoundry/backups/rr_code_backup/__pycache__
- Secret hits: api_key in omni/agentxfoundry/fragments/training_examples/render_workflow_optimization.py.txt, api_key in omni/agentxfoundry/fragments/training_examples/3D_model_training.py.txt, api_key in omni/agentxfoundry/llm_integrations/clients.py
- Notes: Checked-in artifact dirs: 11. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/agentxfoundry/legacy`

- Name: `js-character-generator`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Shell, TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: npm run build
- Artifact dirs: omni/agentxfoundry/legacy/node_modules
- Secret hits: password in omni/agentxfoundry/legacy/Jasteri.sh
- Notes: Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/arch-alive/qenv/lib/python3.14/site-packages/mistralai/extra`

- Name: `extra`
- Status: `watch` (score `65`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Artifact dirs: omni/arch-alive/qenv/lib/python3.14/site-packages/mistralai/extra/__pycache__, omni/arch-alive/qenv/lib/python3.14/site-packages/mistralai/extra/mcp/__pycache__, omni/arch-alive/qenv/lib/python3.14/site-packages/mistralai/extra/observability/__pycache__, omni/arch-alive/qenv/lib/python3.14/site-packages/mistralai/extra/run/__pycache__, omni/arch-alive/qenv/lib/python3.14/site-packages/mistralai/extra/tests/__pycache__
- Secret hits: api_key in omni/arch-alive/qenv/lib/python3.14/site-packages/mistralai/extra/README.md
- Notes: Checked-in artifact dirs: 6. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/cali/jeff/deepfiat`

- Name: `deepfiat`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: make build, make clean
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/deepfiat/backend`

- Name: `fiat-btc-backend-c`
- Status: `watch` (score `80`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript, SQL
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: npm run build
- Secret hits: secret in omni/cali/jeff/deepfiat/backend/src/modules/auth/auth.module.ts
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/cali/jeff/deepfiat/backend/c`

- Name: `c`
- Status: `healthy` (score `100`)
- Kind: `compiled_project`
- Manifests: Makefile
- Docs: none
- Languages: C
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: make, make clean
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/deepfiat/frontend`

- Name: `fiat-btc-frontend`
- Status: `critical` (score `55`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint
- Artifact dirs: omni/cali/jeff/deepfiat/frontend/node_modules
- Secret hits: password in omni/cali/jeff/deepfiat/frontend/src/pages/admin/Login.tsx
- Notes: Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/cali/jeff/deepfiat/zk`

- Name: `fiat-btc-zk-circuit`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: Makefile, package.json
- Docs: none
- Languages: TypeScript, Shell
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: npm run build, make, make build, make test, make clean

### `omni/cali/jeff/theratouch+`

- Name: `theratouch+`
- Status: `watch` (score `62`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: JavaScript, Shell, SQL
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build
- Artifact dirs: omni/cali/jeff/theratouch+/node_modules
- Secret hits: secret in omni/cali/jeff/theratouch+/vendor/pragmarx/google2fa/README.md, password in omni/cali/jeff/theratouch+/vendor/phpunit/php-code-coverage/src/Report/Html/Renderer/Template/js/jquery.min.js, password in omni/cali/jeff/theratouch+/vendor/dflydev/dot-access-data/README.md, password in omni/cali/jeff/theratouch+/vendor/psr/http-message/docs/PSR7-Interfaces.md, password in omni/cali/jeff/theratouch+/vendor/laravel/sail/database/mariadb/create-testing-database.sh
- Notes: Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. Hardcoded workspace paths found: 20. Literal credential-like assignments were detected in text files.

### `omni/cali/jeff/theratouch+/vendor/laravel/framework/src/Illuminate/Foundation/resources/exceptions/renderer`

- Name: `renderer`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build
- Artifact dirs: omni/cali/jeff/theratouch+/vendor/laravel/framework/src/Illuminate/Foundation/resources/exceptions/renderer/dist
- Notes: Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/theratouch+/vendor/mockery/mockery/docs`

- Name: `docs`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: make clean
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar`

- Name: `tri-mar`
- Status: `healthy` (score `87`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: AGENTS.md
- Languages: Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE DATA PREFLIGHT, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Datasets: 1 unique groups, 0 low-fit for raw LLM analysis. Code is present but test coverage surface is thin or not wired into inferred commands. Hardcoded workspace paths found: 24.

### `omni/cali/jeff/tri-mar/platform`

- Name: `platform`
- Status: `critical` (score `40`)
- Kind: `compiled_project`
- Manifests: Makefile
- Docs: README.md
- Languages: C, Python, Shell, Go
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: make, make build, make test, make clean
- Artifact dirs: omni/cali/jeff/tri-mar/platform/vps/__pycache__, omni/cali/jeff/tri-mar/platform/vps/chaos/__pycache__, omni/cali/jeff/tri-mar/platform/vps/crypto/__pycache__, omni/cali/jeff/tri-mar/platform/vps/crypto/hybrid/__pycache__, omni/cali/jeff/tri-mar/platform/vps/crypto/jwt/__pycache__
- Private key paths: omni/cali/jeff/tri-mar/platform/base/cdp_api_key.json
- Secret hits: password in omni/cali/jeff/tri-mar/platform/README.md, password in omni/cali/jeff/tri-mar/platform/infra/helm/fiat-btc/templates/secrets.yaml, secret in omni/cali/jeff/tri-mar/platform/vps/crypto/kem/kyber.c, secret in omni/cali/jeff/tri-mar/platform/vps/crypto/kem/kyber.py, secret in omni/cali/jeff/tri-mar/platform/vps/crypto/kem/kyber.h
- Notes: Checked-in artifact dirs: 16. Private key material appears inside the project tree. Literal credential-like assignments were detected in text files.

### `omni/cali/jeff/tri-mar/platform/backend`

- Name: `fiat-btc-backend`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test, npm run test:coverage
- Artifact dirs: omni/cali/jeff/tri-mar/platform/backend/dist, omni/cali/jeff/tri-mar/platform/backend/node_modules
- Secret hits: password in omni/cali/jeff/tri-mar/platform/backend/src/app.module.ts, password in omni/cali/jeff/tri-mar/platform/backend/src/database/data-source.ts, secret in omni/cali/jeff/tri-mar/platform/backend/src/config/app.config.ts, password in omni/cali/jeff/tri-mar/platform/backend/src/config/app.config.ts, secret in omni/cali/jeff/tri-mar/platform/backend/src/common/interceptors/audit-log.interceptor.ts
- Notes: Checked-in artifact dirs: 2. Literal credential-like assignments were detected in text files.

### `omni/cali/jeff/tri-mar/platform/cmd/fiat-btc`

- Name: `fiat-btc`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: README.md
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/forge-std`

- Name: `forge-std`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts`

- Name: `openzeppelin-solidity`
- Status: `watch` (score `80`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: JavaScript, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER
- Inferred commands: npm run lint, npm run test
- Secret hits: api_key in omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts/scripts/fetch-common-contracts.js
- Notes: Literal credential-like assignments were detected in text files.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable`

- Name: `openzeppelin-solidity`
- Status: `watch` (score `80`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: JavaScript, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER
- Inferred commands: npm run lint, npm run test
- Secret hits: api_key in omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/scripts/fetch-common-contracts.js
- Notes: Literal credential-like assignments were detected in text files.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/contracts`

- Name: `@openzeppelin/contracts-upgradeable`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/fv`

- Name: `fv`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: make clean
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/forge-std`

- Name: `forge-std`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts`

- Name: `openzeppelin-solidity`
- Status: `watch` (score `80`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: JavaScript, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER
- Inferred commands: npm run lint, npm run test
- Secret hits: api_key in omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/scripts/fetch-common-contracts.js
- Notes: Literal credential-like assignments were detected in text files.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/contracts`

- Name: `@openzeppelin/contracts`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/fv`

- Name: `fv`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: make clean
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/lib/forge-std`

- Name: `forge-std`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/scripts/solhint-custom`

- Name: `solhint-plugin-openzeppelin`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts-upgradeable/scripts/solhint-custom`

- Name: `solhint-plugin-openzeppelin`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts/contracts`

- Name: `@openzeppelin/contracts`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts/fv`

- Name: `fv`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: make clean
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts/lib/forge-std`

- Name: `forge-std`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `omni/cali/jeff/tri-mar/platform/contracts/lib/openzeppelin-contracts/scripts/solhint-custom`

- Name: `solhint-plugin-openzeppelin`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `omni/cali/jeff/tri-mar/platform/frontend`

- Name: `fiat-btc-frontend`
- Status: `watch` (score `75`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test, npm run test:coverage
- Artifact dirs: omni/cali/jeff/tri-mar/platform/frontend/.next, omni/cali/jeff/tri-mar/platform/frontend/node_modules
- Notes: Checked-in artifact dirs: 2. Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/vps/network`

- Name: `network`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/vps/network/ebpf`

- Name: `ebpf`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: C, Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/vps/network/tor`

- Name: `tor`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/vps/orchestrator`

- Name: `orchestrator`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/vps/proto`

- Name: `proto`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/cali/jeff/tri-mar/platform/zk`

- Name: `fiat-btc-zk`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Artifact dirs: omni/cali/jeff/tri-mar/platform/zk/node_modules
- Notes: Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/crdroid-source/.repo/repo`

- Name: `repo`
- Status: `watch` (score `70`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: pytest, black --check .
- Artifact dirs: omni/crdroid-source/.repo/repo/__pycache__, omni/crdroid-source/.repo/repo/subcmds/__pycache__
- Secret hits: password in omni/crdroid-source/.repo/repo/main.py, password in omni/crdroid-source/.repo/repo/subcmds/sync.py, password in omni/crdroid-source/.repo/repo/docs/smart-sync.md
- Notes: Checked-in artifact dirs: 2. Literal credential-like assignments were detected in text files.

### `omni/distributed-ai-cluster`

- Name: `ihep-enterprise`
- Status: `critical` (score `32`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: CLAUDE.md
- Languages: Python, Shell, SQL
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE DATA PREFLIGHT
- Inferred commands: pytest, ruff check ., mypy .
- Artifact dirs: omni/distributed-ai-cluster/agent_env, omni/distributed-ai-cluster/agents/predictive/__pycache__, omni/distributed-ai-cluster/enterprise/__pycache__, omni/distributed-ai-cluster/enterprise/approval/__pycache__, omni/distributed-ai-cluster/enterprise/audit/__pycache__
- Private key paths: omni/distributed-ai-cluster/google-credentials.json
- Secret hits: PGPASSWORD in omni/distributed-ai-cluster/PROCEDURAL_REGISTRY_GUIDE.md, api_key in omni/distributed-ai-cluster/CODE_CONSOLIDATION_DEMO.md, api_key in omni/distributed-ai-cluster/K3S_CLUSTER.md, PGPASSWORD in omni/distributed-ai-cluster/PROCEDURAL_REGISTRY_IMPLEMENTATION.md, PGPASSWORD in omni/distributed-ai-cluster/SESSION_HANDOFF.md
- Notes: Checked-in artifact dirs: 21. Datasets: 1 unique groups, 0 low-fit for raw LLM analysis. Contains contact-style data fields; redact before external model use. Hardcoded workspace paths found: 9. Private key material appears inside the project tree. Literal credential-like assignments were detected in text files.

### `omni/distributed-ai-cluster/dashboard`

- Name: `dashboard`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript, SQL, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: npm run dev, npm run build, npm run lint, npm run type-check, npm run test, npm run test:coverage
- Artifact dirs: omni/distributed-ai-cluster/dashboard/.next, omni/distributed-ai-cluster/dashboard/node_modules
- Secret hits: password in omni/distributed-ai-cluster/dashboard/src/lib/db.ts
- Notes: Checked-in artifact dirs: 2. Literal credential-like assignments were detected in text files.

### `omni/distributed-ai-cluster/legacy`

- Name: `js-character-generator`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Shell, TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: npm run build
- Artifact dirs: omni/distributed-ai-cluster/legacy/node_modules
- Secret hits: password in omni/distributed-ai-cluster/legacy/Jasteri.sh
- Notes: Checked-in artifact dirs: 1. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/eight-layer-pqc`

- Name: `eight-layer-pqc`
- Status: `watch` (score `60`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: README.md
- Languages: Python, Shell, JavaScript, SQL
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE DATA PREFLIGHT, THE DUPLICATE TREE HUNTER
- Inferred commands: pytest, black --check ., mypy .
- Artifact dirs: omni/eight-layer-pqc/.venv, omni/eight-layer-pqc/venv
- Secret hits: secret in omni/eight-layer-pqc/code/python/layer4_encryption.py
- Notes: Checked-in artifact dirs: 2. Datasets: 5 unique groups, 0 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 5. Literal credential-like assignments were detected in text files.

### `omni/eight-layer-pqc/code/go`

- Name: `go`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/eight-layer-pqc/code/rust`

- Name: `rust`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Cargo.toml
- Docs: none
- Languages: Rust
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights`

- Name: `health-insights`
- Status: `critical` (score `5`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: AGENTS.md, CLAUDE.md, README.md
- Languages: TypeScript, Python, Shell, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE DATA PREFLIGHT, THE DUPLICATE TREE HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Artifact dirs: omni/health-insights/.next, omni/health-insights/.pytest_cache, omni/health-insights/.venv, omni/health-insights/node_modules, omni/health-insights/venv
- Private key paths: omni/health-insights/docs/ihep-app.2025-12-01.private-key.pem
- Secret hits: PGPASSWORD in omni/health-insights/setup.sh, password in omni/health-insights/python/ihep-auth-api-service.py, secret in omni/health-insights/markdown/IHEP_Phase_III_Security_Architecture copy.md, password in omni/health-insights/markdown/IHEP_Phase_III_Security_Architecture copy.md, secret in omni/health-insights/markdown/IHEP_Phase_III_Security_Architecture.md
- Notes: Checked-in artifact dirs: 5. Datasets: 22 unique groups, 11 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 21. Code is present but test coverage surface is thin or not wired into inferred commands. Contains placeholder or synthetic training data that should be cleaned before LLM use. Contains patient-profile fields; treat the data as sensitive. Private key material appears inside the project tree. Literal credential-like assignments were detected in text files.

### `omni/health-insights/app_backup`

- Name: `ihep-application`
- Status: `watch` (score `65`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript, Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test
- Secret hits: password in omni/health-insights/app_backup/auth/signup/page.tsx, password in omni/health-insights/app_backup/auth/login/page.tsx, password in omni/health-insights/app_backup/backend/shared/utils/validation.py, password in omni/health-insights/app_backup/backend/auth-service/app.py
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/health-insights/applications`

- Name: `@ihep/frontend`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Python, SQL, TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: npm run dev, npm run build, npm run lint, npm run type-check, npm run test, npm run test:coverage
- Artifact dirs: omni/health-insights/applications/backend/.pytest_cache, omni/health-insights/applications/backend/.venv-security, omni/health-insights/applications/backend/node_modules, omni/health-insights/applications/backend/venv
- Secret hits: password in omni/health-insights/applications/backend/app.py, password in omni/health-insights/applications/backend/shared/utils/validation.py, password in omni/health-insights/applications/backend/auth-service/app.py
- Notes: Checked-in artifact dirs: 4. Literal credential-like assignments were detected in text files.

### `omni/health-insights/applications/frontend`

- Name: `@ihep/frontend`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint, npm run type-check, npm run test, npm run test:coverage
- Artifact dirs: omni/health-insights/applications/frontend/.next, omni/health-insights/applications/frontend/node_modules
- Secret hits: password in omni/health-insights/applications/frontend/src/components/auth/AuthProvider.tsx
- Notes: Checked-in artifact dirs: 2. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/health-insights/ihep-application`

- Name: `ihep-platform`
- Status: `critical` (score `40`)
- Kind: `node_app`
- Manifests: package.json
- Docs: AGENTS.md, CLAUDE.md, README.md
- Languages: TypeScript, Python, Shell, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE DATA PREFLIGHT, THE DUPLICATE TREE HUNTER
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test, npm run test:coverage
- Artifact dirs: omni/health-insights/ihep-application/.next, omni/health-insights/ihep-application/node_modules, omni/health-insights/ihep-application/services/clinical-twin-service/__pycache__
- Secret hits: PGPASSWORD in omni/health-insights/ihep-application/setup.sh, password in omni/health-insights/ihep-application/MOCK_DATA_REPLACEMENT_SUMMARY.md, password in omni/health-insights/ihep-application/migrations/001_users_providers.sql, secret in omni/health-insights/ihep-application/docs/SECURITY_STATUS_FINAL.md, api_key in omni/health-insights/ihep-application/docs/INFRASTRUCTURE_CONNECTIONS.md
- Notes: Checked-in artifact dirs: 3. Datasets: 21 unique groups, 11 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 21. Contains placeholder or synthetic training data that should be cleaned before LLM use. Contains patient-profile fields; treat the data as sensitive. Literal credential-like assignments were detected in text files.

### `omni/health-insights/ihep-application/hub/gateway`

- Name: `@ihep/frontend`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Python, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint, npm run type-check, npm run test, npm run test:coverage
- Artifact dirs: omni/health-insights/ihep-application/hub/gateway/ehr/__pycache__, omni/health-insights/ihep-application/hub/gateway/ehr/sync/__pycache__, omni/health-insights/ihep-application/hub/gateway/ehr/transformers/__pycache__, omni/health-insights/ihep-application/hub/gateway/ehr/webhooks/__pycache__
- Secret hits: api_key in omni/health-insights/ihep-application/hub/gateway/ehr/config/ehr-partners/allscripts-example.yaml, api_key in omni/health-insights/ihep-application/hub/gateway/ehr/adapters/allscripts_adapter.py, api_key in omni/health-insights/ihep-application/hub/gateway/ehr/adapters/base_adapter.py, secret in omni/health-insights/ihep-application/hub/gateway/ehr/webhooks/handler.py
- Notes: Checked-in artifact dirs: 4. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/health-insights/terraform/.terraform/modules/autokey-org-policy-gcp-restrict-cmek-crypto-key-projects`

- Name: `autokey-org-policy-gcp-restrict-cmek-crypto-key-projects`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/health-insights/terraform/.terraform/modules/autokey-org-policy-gcp-restrict-cmek-crypto-key-projects/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/terraform/.terraform/modules/cs-autokey`

- Name: `cs-autokey`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/terraform/.terraform/modules/cs-autokey/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/terraform/.terraform/modules/cs-common`

- Name: `cs-common`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/terraform/.terraform/modules/cs-folders-iam-0-computeinstanceAdminv1`

- Name: `cs-folders-iam-0-computeinstanceAdminv1`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/health-insights/terraform/.terraform/modules/cs-folders-iam-0-computeinstanceAdminv1/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/health-insights/terraform/.terraform/modules/cs-gg-lifescien-concept-nonprod-svc`

- Name: `cs-gg-lifescien-concept-nonprod-svc`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/health-insights/terraform/.terraform/modules/cs-kms-projects`

- Name: `cs-kms-projects`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Python, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/terraform/.terraform/modules/cs-kms-projects/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/terraform/.terraform/modules/cs-logging-destination`

- Name: `cs-logging-destination`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: SQL
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/terraform/.terraform/modules/cs-logging-destination/modules/bq-log-alerting/logging/cloud_function`

- Name: `cloud_function`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `omni/health-insights/terraform/.terraform/modules/cs-logging-destination/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/terraform/.terraform/modules/cs-vpc-dev-shared`

- Name: `cs-vpc-dev-shared`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/terraform/.terraform/modules/cs-vpc-dev-shared/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/health-insights/workspaces`

- Name: `ihep-application`
- Status: `watch` (score `80`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Python, TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test
- Secret hits: password in omni/health-insights/workspaces/backend/shared/utils/validation.py, password in omni/health-insights/workspaces/backend/auth-service/app.py
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/health-insights/workspaces/frontend`

- Name: `@ihep/frontend`
- Status: `watch` (score `80`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint, npm run type-check, npm run test, npm run test:coverage
- Secret hits: password in omni/health-insights/workspaces/frontend/src/components/auth/AuthProvider.tsx
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/ihep-app`

- Name: `ihep-app`
- Status: `critical` (score `45`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: C, Shell, JavaScript, Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Artifact dirs: omni/ihep-app/.pytest_cache, omni/ihep-app/.venv, omni/ihep-app/.venv-audit, omni/ihep-app/actions-runner/externals/node20/lib/node_modules, omni/ihep-app/actions-runner/externals/node24/lib/node_modules
- Secret hits: api_key in omni/ihep-app/ihep-fintech-training.py.txt, password in omni/ihep-app/pu-to-docker.txt, secret in omni/ihep-app/markdown/IHEP_Phase_III_Security_Architecture.md, password in omni/ihep-app/markdown/IHEP_Phase_III_Security_Architecture.md, secret in omni/ihep-app/markdown/IHEP_Communication_NIST_Control_Mapping.md
- Notes: Checked-in artifact dirs: 6. Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts. Literal credential-like assignments were detected in text files.

### `omni/ihep-app/ihep`

- Name: `ihep-web`
- Status: `critical` (score `40`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript, Python, Shell, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE DATA PREFLIGHT, THE DUPLICATE TREE HUNTER
- Inferred commands: npm run dev, npm run build, npm run test
- Artifact dirs: omni/ihep-app/ihep/.next, omni/ihep-app/ihep/.pytest_cache, omni/ihep-app/ihep/.venv, omni/ihep-app/ihep/models/medgemma-4b-it/venv, omni/ihep-app/ihep/node_modules
- Secret hits: PGPASSWORD in omni/ihep-app/ihep/setup.sh, password in omni/ihep-app/ihep/python/ihep-auth-api-service.py, secret in omni/ihep-app/ihep/markdown/IHEP_Phase_III_Security_Architecture copy.md, password in omni/ihep-app/ihep/markdown/IHEP_Phase_III_Security_Architecture copy.md, secret in omni/ihep-app/ihep/markdown/IHEP_Phase_III_Security_Architecture.md
- Notes: Checked-in artifact dirs: 5. Datasets: 21 unique groups, 11 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 21. Contains placeholder or synthetic training data that should be cleaned before LLM use. Contains patient-profile fields; treat the data as sensitive. Literal credential-like assignments were detected in text files.

### `omni/ihep-app/ihep/app`

- Name: `ihep-application`
- Status: `watch` (score `80`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Python, TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint, npm run test
- Secret hits: password in omni/ihep-app/ihep/app/backend/shared/utils/validation.py, password in omni/ihep-app/ihep/app/backend/auth-service/app.py
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/ihep-app/ihep/app/frontend`

- Name: `@ihep/frontend`
- Status: `watch` (score `80`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint, npm run type-check, npm run test, npm run test:coverage
- Secret hits: password in omni/ihep-app/ihep/app/frontend/src/components/auth/AuthProvider.tsx
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/ihep-app/ihep/applications`

- Name: `@ihep/frontend`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: Python, SQL, TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: npm run dev, npm run build, npm run lint, npm run type-check, npm run test, npm run test:coverage
- Artifact dirs: omni/ihep-app/ihep/applications/backend/.pytest_cache, omni/ihep-app/ihep/applications/backend/.venv-security, omni/ihep-app/ihep/applications/backend/__pycache__, omni/ihep-app/ihep/applications/backend/node_modules, omni/ihep-app/ihep/applications/backend/venv
- Secret hits: password in omni/ihep-app/ihep/applications/backend/app.py, password in omni/ihep-app/ihep/applications/backend/shared/utils/validation.py, password in omni/ihep-app/ihep/applications/backend/auth-service/app.py
- Notes: Checked-in artifact dirs: 6. Literal credential-like assignments were detected in text files.

### `omni/ihep-app/ihep/applications/frontend`

- Name: `@ihep/frontend`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint, npm run type-check, npm run test, npm run test:coverage
- Artifact dirs: omni/ihep-app/ihep/applications/frontend/.next, omni/ihep-app/ihep/applications/frontend/node_modules
- Secret hits: password in omni/ihep-app/ihep/applications/frontend/src/components/auth/AuthProvider.tsx
- Notes: Checked-in artifact dirs: 2. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/ihep-app/ihep/terraform/.terraform/modules/autokey-org-policy-gcp-restrict-cmek-crypto-key-projects`

- Name: `autokey-org-policy-gcp-restrict-cmek-crypto-key-projects`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/ihep-app/ihep/terraform/.terraform/modules/autokey-org-policy-gcp-restrict-cmek-crypto-key-projects/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-autokey`

- Name: `cs-autokey`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-autokey/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-common`

- Name: `cs-common`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-folders-iam-0-computeinstanceAdminv1`

- Name: `cs-folders-iam-0-computeinstanceAdminv1`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-folders-iam-0-computeinstanceAdminv1/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-gg-lifescien-concept-nonprod-svc`

- Name: `cs-gg-lifescien-concept-nonprod-svc`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: none detected
- Workflow stack: THE PROJECT RECON BLITZ
- Inferred commands: none inferred

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-kms-projects`

- Name: `cs-kms-projects`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Python, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-kms-projects/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-logging-destination`

- Name: `cs-logging-destination`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: SQL
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-logging-destination/modules/bq-log-alerting/logging/cloud_function`

- Name: `cloud_function`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-logging-destination/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-vpc-dev-shared`

- Name: `cs-vpc-dev-shared`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Makefile
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/ihep-app/ihep/terraform/.terraform/modules/cs-vpc-dev-shared/test/integration`

- Name: `integration`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/jarmacz.com`

- Name: `jarmacz-com`
- Status: `critical` (score `55`)
- Kind: `node_app`
- Manifests: package.json
- Docs: AGENTS.md, README.md
- Languages: TypeScript, JavaScript, Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE DATA PREFLIGHT, THE BUILD TRUTH SERUM
- Inferred commands: npm run dev, npm run build, npm run lint
- Artifact dirs: omni/jarmacz.com/.next, omni/jarmacz.com/node_modules, omni/jarmacz.com/path/to/venv
- Secret hits: password in omni/jarmacz.com/out/_next/static/chunks/249261e921aeebba.js, password in omni/jarmacz.com/out/_next/static/chunks/a6dad97d9634a72d.js, password in omni/jarmacz.com/api/app.py
- Notes: Checked-in artifact dirs: 3. Datasets: 1 unique groups, 0 low-fit for raw LLM analysis. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/multimodal-agent-builder`

- Name: `multimodal-agent-builder`
- Status: `watch` (score `62`)
- Kind: `mixed_node_python`
- Manifests: package.json, pyproject.toml
- Docs: AGENTS.md, README.md
- Languages: TypeScript, Python, Shell, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE DATA PREFLIGHT
- Inferred commands: pytest, ruff check ., black --check ., mypy .
- Artifact dirs: omni/multimodal-agent-builder/.venv, omni/multimodal-agent-builder/config/__pycache__, omni/multimodal-agent-builder/node_modules, omni/multimodal-agent-builder/src/__pycache__, omni/multimodal-agent-builder/src/agents/__pycache__
- Secret hits: password in omni/multimodal-agent-builder/NEURO_INSIGHT_ENGINE_REMOTE_BUILD_SPEC.md, api_key in omni/multimodal-agent-builder/tests/conftest.py, api_key in omni/multimodal-agent-builder/tests/unit/test_llm_clients_basic.py, api_key in omni/multimodal-agent-builder/tests/unit/test_llm_clients_old.py, api_key in omni/multimodal-agent-builder/tests/unit/test_llm_clients.py
- Notes: Checked-in artifact dirs: 8. Datasets: 17 unique groups, 2 low-fit for raw LLM analysis. Literal credential-like assignments were detected in text files.

### `omni/neurodivergence.works/security-architecture/eight-layer-pqc`

- Name: `eight-layer-pqc`
- Status: `watch` (score `65`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: README.md
- Languages: Python, Shell, SQL
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE DATA PREFLIGHT, THE DUPLICATE TREE HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Secret hits: secret in omni/neurodivergence.works/security-architecture/eight-layer-pqc/code/python/layer4_encryption.py
- Notes: Datasets: 5 unique groups, 0 low-fit for raw LLM analysis. Cross-project duplicate dataset groups: 5. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `omni/neurodivergence.works/security-architecture/eight-layer-pqc/code/go`

- Name: `go`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: none
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/neurodivergence.works/security-architecture/eight-layer-pqc/code/rust`

- Name: `rust`
- Status: `healthy` (score `100`)
- Kind: `generic_build_project`
- Manifests: Cargo.toml
- Docs: none
- Languages: Rust
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `omni/provenance-engine`

- Name: `provenance-engine`
- Status: `healthy` (score `90`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: README.md
- Languages: Shell, Python
- Workflow stack: THE PROJECT RECON BLITZ, THE ARTIFACT PURGE
- Inferred commands: pytest
- Artifact dirs: omni/provenance-engine/__pycache__, omni/provenance-engine/dist, omni/provenance-engine/venv
- Notes: Checked-in artifact dirs: 3.

### `omni/provenance-engine/build/pkg/payload/opt/provenance-engine`

- Name: `provenance-engine`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: README.md
- Languages: Shell, Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `omni/security-workspace`

- Name: `security-workspace`
- Status: `watch` (score `65`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: README.md
- Languages: Python, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE DATA PREFLIGHT, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Artifact dirs: omni/security-workspace/__pycache__
- Secret hits: api_key in omni/security-workspace/security_convergence.py
- Notes: Checked-in artifact dirs: 1. Datasets: 1 unique groups, 0 low-fit for raw LLM analysis. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `tools/foundational-frameworks`

- Name: `@tools/foundational-frameworks`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE ARTIFACT PURGE
- Inferred commands: npm run build, npm run lint, npm run test
- Artifact dirs: tools/foundational-frameworks/dist, tools/foundational-frameworks/node_modules
- Notes: Checked-in artifact dirs: 2.

### `tools/gemini-cli-0.22.2`

- Name: `@google/gemini-cli`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: Makefile, package.json
- Docs: README.md
- Languages: TypeScript, JavaScript, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: npm run build, npm run lint, npm run test, make build, make test, make clean
- Artifact dirs: tools/gemini-cli-0.22.2/node_modules
- Secret hits: password in tools/gemini-cli-0.22.2/.github/workflows/gemini-scheduled-issue-dedup.yml, password in tools/gemini-cli-0.22.2/.github/workflows/gemini-automated-issue-dedup.yml, secret in tools/gemini-cli-0.22.2/.github/workflows/chained_e2e.yml, secret in tools/gemini-cli-0.22.2/.github/workflows/release-sandbox.yml, secret in tools/gemini-cli-0.22.2/.github/workflows/ci.yml
- Notes: Checked-in artifact dirs: 1. Literal credential-like assignments were detected in text files.

### `tools/gemini-cli-0.22.2/packages/a2a-server`

- Name: `@google/gemini-cli-a2a-server`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE ARTIFACT PURGE
- Inferred commands: npm run build, npm run lint, npm run test
- Artifact dirs: tools/gemini-cli-0.22.2/packages/a2a-server/dist, tools/gemini-cli-0.22.2/packages/a2a-server/node_modules
- Notes: Checked-in artifact dirs: 2.

### `tools/gemini-cli-0.22.2/packages/cli`

- Name: `@google/gemini-cli`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: npm run build, npm run lint, npm run test
- Artifact dirs: tools/gemini-cli-0.22.2/packages/cli/node_modules
- Secret hits: secret in tools/gemini-cli-0.22.2/packages/cli/src/config/extensions/extensionSettings.ts
- Notes: Checked-in artifact dirs: 1. Literal credential-like assignments were detected in text files.

### `tools/gemini-cli-0.22.2/packages/cli/src/commands/extensions/examples/mcp-server`

- Name: `mcp-server-example`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run build
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `tools/gemini-cli-0.22.2/packages/core`

- Name: `@google/gemini-cli-core`
- Status: `critical` (score `40`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: npm run build, npm run lint, npm run test
- Artifact dirs: tools/gemini-cli-0.22.2/packages/core/node_modules
- Private key paths: tools/gemini-cli-0.22.2/packages/core/src/telemetry/sdk.test.ts
- Secret hits: password in tools/gemini-cli-0.22.2/packages/core/src/mcp/token-storage/keychain-token-storage.test.ts, password in tools/gemini-cli-0.22.2/packages/core/src/mcp/token-storage/keychain-token-storage.ts, api_key in tools/gemini-cli-0.22.2/packages/core/src/telemetry/sanitize.ts, api_key in tools/gemini-cli-0.22.2/packages/core/src/telemetry/sanitize.test.ts, api_key in tools/gemini-cli-0.22.2/packages/core/src/telemetry/metrics.test.ts
- Notes: Checked-in artifact dirs: 1. Private key material appears inside the project tree. Literal credential-like assignments were detected in text files.

### `tools/gemini-cli-0.22.2/packages/test-utils`

- Name: `@google/gemini-cli-test-utils`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: TypeScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run build
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `tools/gemini-cli-0.22.2/packages/vscode-ide-companion`

- Name: `gemini-cli-vscode-ide-companion`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: README.md
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE ARTIFACT PURGE
- Inferred commands: npm run build, npm run lint, npm run test
- Artifact dirs: tools/gemini-cli-0.22.2/packages/vscode-ide-companion/node_modules
- Notes: Checked-in artifact dirs: 1.

### `tools/gemini-cli-0.22.2/third_party/get-ripgrep`

- Name: `@lvce-editor/ripgrep`
- Status: `healthy` (score `100`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: npm run test
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands.

### `tools/github-mcp-server`

- Name: `github-mcp-server`
- Status: `watch` (score `80`)
- Kind: `generic_build_project`
- Manifests: go.mod
- Docs: README.md
- Languages: Go
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Secret hits: password in tools/github-mcp-server/.github/workflows/docker-publish.yml
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

### `tools/mcp-server`

- Name: `mcp-server`
- Status: `watch` (score `80`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER
- Inferred commands: pytest
- Secret hits: password in tools/mcp-server/.github/workflows/push_to_docker_hub.yml
- Notes: Literal credential-like assignments were detected in text files.

### `tools/mcpb`

- Name: `@anthropic-ai/mcpb`
- Status: `watch` (score `70`)
- Kind: `node_app`
- Manifests: package.json
- Docs: CLAUDE.md, README.md
- Languages: TypeScript, JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE
- Inferred commands: yarn dev, yarn build, yarn lint, yarn test
- Artifact dirs: tools/mcpb/dist, tools/mcpb/node_modules
- Secret hits: api_key in tools/mcpb/test/schemas.test.ts, api_key in tools/mcpb/test/init.test.ts
- Notes: Checked-in artifact dirs: 2. Literal credential-like assignments were detected in text files.

### `tools/mcpb/examples/file-manager-python`

- Name: `file-manager-python`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: none
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `tools/mcpb/examples/file-system-node`

- Name: `ant.dir.ant.anthropic.filesystem`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `tools/mcpb/examples/hello-world-node`

- Name: `hello-world-node`
- Status: `healthy` (score `90`)
- Kind: `node_app`
- Manifests: package.json
- Docs: none
- Languages: JavaScript
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. package.json exists but defines no scripts.

### `tools/mcpb/examples/hello-world-uv`

- Name: `hello-world-uv`
- Status: `healthy` (score `100`)
- Kind: `python_project`
- Manifests: pyproject.toml
- Docs: README.md
- Languages: Python
- Workflow stack: THE PROJECT RECON BLITZ, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Notes: Code is present but test coverage surface is thin or not wired into inferred commands. pyproject exists but no test files were detected in the owned tree.

### `tools/skills`

- Name: `skills`
- Status: `watch` (score `65`)
- Kind: `docs_or_code_project`
- Manifests: none
- Docs: CLAUDE.md, README.md
- Languages: Python, JavaScript, Shell
- Workflow stack: THE PROJECT RECON BLITZ, THE SECRET HUNTER, THE ARTIFACT PURGE, THE BUILD TRUTH SERUM
- Inferred commands: none inferred
- Artifact dirs: tools/skills/dist, tools/skills/skills/skill-creator/scripts/__pycache__
- Secret hits: password in tools/skills/skills/pdf/reference.md, password in tools/skills/skills/pdf/SKILL.md, api_key in tools/skills/skills/mcp-builder/reference/python_mcp_server.md
- Notes: Checked-in artifact dirs: 2. Code is present but test coverage surface is thin or not wired into inferred commands. Literal credential-like assignments were detected in text files.

