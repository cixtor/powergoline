# PowerGoLine

A lightweight status line for your terminal emulator. This project aims to be a lightweight alternative for the [powerline](https://github.com/powerline/powerline) project which is a status line plugin for vim that also provides support for several other applications, including zsh, bash, tmux, IPython, Awesome and Qtile.

![PowerGoLine Screenshot](screenshot.png)

### Installation

1. Clone this repository in your computer
2. Install a patched monospace font [from here](https://github.com/powerline/fonts)
3. Build the binary with `go build` _(not extra dependencies)_
4. Place the binary in your home directory
5. If you are not using _Bash_ [use this](https://github.com/milkbikis/powerline-shell) instead
6. Add this function to your `.bashrc` configuration

```
function set_prompt_command() {
    export PS1="$($HOME/powergoline $? 2> /dev/null)"
}
export PROMPT_COMMAND="set_prompt_command; $PROMPT_COMMAND"
```

### Configuration

The binary places a JSON file with the default configuration in the home directory of the user in the current session named `.powergoline.json`, change the values available in that file and leave the default values of the unchanged keys, if you want to reset the configuration just delete this file and the binary will re-create its content automatically.

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
* Branch and status info for Mercurial repository.

### TODO

* Display current CVS branch _(git, subversion)_.
* Display number of open processes _(all users, current user)_.
* Support extensions for extra information.

### License

```
The MIT License (MIT)

Copyright (c) 2015 CIXTOR

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
