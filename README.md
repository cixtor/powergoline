# Powergoline

A lightweight status line for your terminal emulator. This project aims to be a lightweight alternative for [powerline](https://github.com/powerline/powerline) a popular statusline plugin for VIm that statuslines and prompts for several other applications, including zsh, bash, tmux, IPython, Awesome and Qtile.

## Installation

1. Install a patched monospace font [from here](https://github.com/powerline/fonts)
1. `go install github.com/cixtor/powergoline@latest`
1. Add this function to your `.bashrc` configuration
   ```sh
   function set_prompt_command() {
     RESULT=$(powergoline -theme="wildcherry" -status.code="$?")
     export PS1="$RESULT"
   }
   export PROMPT_COMMAND="set_prompt_command; $PROMPT_COMMAND"
   ```
1. Restart your terminal emulator.

![powergoline](screenshot.png)

## Configuration

Use `powergoline -h` to see all available options.

Update the `set_prompt_command` function to add or remove flags accordingly. 

Select a predefined color scheme using the `-theme` flag and one of these values: agnoster, astrocom, bluescale, colorish, grayscale, wildcherry, or create your own by passing the corresponding `-ABC.fg` and `-ABC.bg` flags for the foreground and background colors, respectivevly.

## Plugins

Add one or more `-plugin="..."` flags to `set_prompt_command`.

Each plugin must execute a command available in `$PATH`.

Background and foreground colors are automatically selected based on the surrouding prompt segments.

Report errors via `/dev/stderr` and stop the program with `exit(1)` in your corresponding language.

# Performance

Average performance with the default features:

```sh
$ hyperfine --shell=none 'powergoline'
Benchmark 1: powergoline
  Time (mean ± σ):      21.1 ms ±   3.8 ms    [User: 8.5 ms, System: 9.6 ms]
  Range (min … max):    15.8 ms …  33.4 ms    109 runs
```

Average performance with the most basic features enabled:

```sh
$ hyperfine --shell=none 'powergoline ...'
Benchmark 1: powergoline -time.on -user.on -host.on -cwd.on -cwd.n=3 -status.code=0
  Time (mean ± σ):       6.8 ms ±   1.8 ms    [User: 2.5 ms, System: 2.3 ms]
  Range (min … max):     4.9 ms …  20.2 ms    343 runs
```

Average performance with the repository feature enabled:

```sh
$ hyperfine --shell=none 'powergoline ...'
Benchmark 1: powergoline -time.on -user.on -host.on -cwd.on -cwd.n=3 -repo.on -status.code=0
  Time (mean ± σ):      21.9 ms ±   4.3 ms    [User: 8.7 ms, System: 10.2 ms]
  Range (min … max):    15.4 ms …  34.8 ms    99 runs
```
