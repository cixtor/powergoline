/**
 * PowerGoLine
 * https://cixtor.com/
 * https://github.com/cixtor/powergoline
 * https://en.wikipedia.org/wiki/Status_bar
 *
 * A status bar is a graphical control element which poses an information area
 * typically found at the window's bottom. It can be divided into sections to
 * group information. Its job is primarily to display information about the
 * current state of its window, although some status bars have extra
 * functionality.
 *
 * A status bar can also be text-based, primarily in console-based applications,
 * in which case it is usually the last row in an 80x25 text mode configuration,
 * leaving the top 24 rows for application data. Usually the status bar (called
 * a status line in this context) displays the current state of the application,
 * as well as helpful keyboard shortcuts.
 */

package main

import (
	"flag"
	"os"
)

// enabled defines a positive status.
const enabled string = "enabled"

// filename is the name of the configuration file.
const filename = ".powergoline.json"

func main() {
	flag.Parse()

	pogol := NewPowerGoLine(os.Getenv("HOME") + "/" + filename)

	pogol.TermTitle()
	pogol.DateTime()
	pogol.Username()
	pogol.Hostname()
	pogol.WorkingDirectory()
	pogol.GitInformation()
	pogol.MercurialInformation()
	pogol.ExecuteAllPlugins()
	pogol.RootSymbol(flag.Arg(0))
	pogol.PrintStatusLine()

	os.Exit(0)
}
