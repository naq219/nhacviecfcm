# Go - Console Commands

## Register Custom Commands

Register custom console commands using `app.RootCmd.AddCommand(cmd)` where `cmd` is a cobra command.

### Basic Example
```go
package main

import (
    "log"
    "github.com/pocketbase/pocketbase"
    "github.com/spf13/cobra"
)

func main() {
    app := pocketbase.New()
    
    app.RootCmd.AddCommand(&cobra.Command{
        Use: "hello",
        Run: func(cmd *cobra.Command, args []string) {
            log.Println("Hello world!")
        },
    })
    
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Running Commands
```bash
# Build and run
./myapp hello

# Or run directly
go run main.go hello
```

## Important Notes
- Console commands execute in their own separate app process
- Commands run independently from the main `serve` command
- Hook and realtime events between different processes are not shared