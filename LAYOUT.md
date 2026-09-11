LAYOUT.md: this ledger
cmd/volatile-downloads/main.go: CLI composition root — flags, config, observe/decide/execute
internal/ensure/plan.go: Observation, Plan, and the pure Decide policy
internal/ensure/ensure.go: observe the live folders, execute a plan
internal/files/files.go: filesystem facade over os
launchd/com.zbornheimer.volatile-downloads.plist: login LaunchAgent pointing at ~/.local/bin
