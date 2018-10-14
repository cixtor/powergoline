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

The binary places a JSON file with the default configuration in the home directory of the user in the current session named `.powergoline.json`, change the values available in that file and leave the default values of the unchanged keys, if you want to reset the configuration just delete this file and the binary will re-create its content automatically.

### Plugins

The binary allows you to execute external commands to complement the information already provided by the built-in features. The output of these external programs is expected to be a single line to keep the format in shape. They will all appear before the status symbol and you can configure the background and foreground colors. You can execute as many plugins as you want. Bellow is an example of a configuration that will execute two external programs:

```
"plugins": [
  {
    "command": "timestamp",
    "background": "255",
    "foreground": "023"
  },
  {
    "command": "shrug",
    "background": "250",
    "foreground": "020"
  }
]
```

### Features

* Color basic status code of executed commands.
* Shorten long directory paths for readability.
* Set terminal title with current directory path.
* Display lock if current directory is read-only.
* Change prompt symbol if super user session.
* No follow symbolic links for brevity.
* Display current time if enabled by user.
* Switch to hide the username segment.
* Switch to hide the hostname segment.
* Branch and status info for Git repository.
* Branch and status info for Mercurial repository.
* Support plugins for extra information.

### TODO

* Display current CVS branch _(subversion)_.
* Display number of open processes _(all users, current user)_.
