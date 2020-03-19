# Powergoline

A lightweight status line for your terminal emulator. This project aims to be a lightweight alternative for [powerline](https://github.com/powerline/powerline) a popular statusline plugin for VIm that statuslines and prompts for several other applications, including zsh, bash, tmux, IPython, Awesome and Qtile.

## Installation

1. Install a patched monospace font [from here](https://github.com/powerline/fonts)
1. `go get -u github.com/cixtor/powergoline`
1. Test using this command: `command powergoline`
1. Add this function to your `.bashrc` configuration

```sh
function set_prompt_command() {
  export PS1="$(powergoline $? 2> /dev/null)"
}
export PROMPT_COMMAND="set_prompt_command; $PROMPT_COMMAND"
```

![powergoline](screenshot.png)

## Configuration

The program creates a JSON file with the default configuration in the home directory `$HOME/.powergoline.json`. You can modify this file to enable or disable different segments of the user prompt. You can control the CVS integration (Git and Mercurial), number of folders, date and time, username, hostname, you can even add your own segments to the prompt with custom plugins. You can reset the configuration by deleting this file.

Some themes are available in `$GOPATH/src/github.com/cixtor/powergoline/themes/`

## Plugins

The program allows you to execute external commands to complement the information provided by the built-in features. The output of these external programs is expected to be a single line. They will all appear before the status symbol and you can configure the background and foreground colors. You can execute as many plugins as you want.

Below is an example using two plugins `timestamp` and `shrug`. Executable files in `$PATH` are also available inside the plugins environment. By default, the prompt prints everything the plugin sends to `/dev/stdout`. You can report errors by sending a message to `/dev/stderr` and stopping the script with exit code `1` (one).

```
"plugins": [
  {
    "command": "/usr/local/bin/timestamp",
    "background": "255",
    "foreground": "023"
  },
  {
    "command": "/Users/foo/.bin/shrug",
    "background": "250",
    "foreground": "020"
  }
]
```

## Performance

```
BenchmarkAll-4                 115     9574368 ns/op
BenchmarkTermTitle-4       3126940         366 ns/op
BenchmarkDatetime-4         947794        1122 ns/op
BenchmarkUsername-4        1444171         789 ns/op
BenchmarkHostname-4        1585209         755 ns/op
BenchmarkDirectories-4      106413       11092 ns/op
BenchmarkRepoStatus-4          136    10050039 ns/op
BenchmarkCallPlugins-4         271     4261030 ns/op
BenchmarkRootSymbol-4      1127618        1169 ns/op
```
