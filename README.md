# Powergoline

A lightweight status line for your terminal emulator. This project aims to be a lightweight alternative for the [powerline](https://github.com/powerline/powerline) project which is a status line plugin for vim that also provides support for several other applications, including zsh, bash, tmux, IPython, Awesome and Qtile.

### Installation

1. Install a patched monospace font [from here](https://github.com/powerline/fonts)
1. `go get -u github.com/cixtor/powergoline`
1. Add this function to your `.bashrc` configuration

```sh
function set_prompt_command() {
  export PS1="$($HOME/powergoline $? 2> /dev/null)"
}
export PROMPT_COMMAND="set_prompt_command; $PROMPT_COMMAND"
```

![powergoline](screenshot.png)

### Configuration

The program creates a JSON file with the default configuration in the home directory `$HOME/.powergoline.json`. You can modify this file to hide different segments of the user prompt. You can control the CVS (Git and Mercurial) integration, the number of folders to display, hide the date, username, or hostname, extend the prompt with custom plugins, among other things. The configuration file resets if you delete this file.

Some themes are available in `$GOPATH/src/github.com/cixtor/powergoline/themes/`

### Plugins

The program allows you to execute external commands to complement the information already provided by the built-in features. The output of these external programs is expected to be a single line. They will all appear before the status symbol, and you can configure the background and foreground colors. You can execute as many plugins as you want.

Below is an example with two plugins `timestamp` and `shrug`. Any command available via `$PATH` is also available inside the plugins environment. Error messages sent to `/dev/stderr` are printed as long as the exit code is one, otherwise, `/dev/stdout` will be preferred.

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

### Performance

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
