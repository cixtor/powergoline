/**
 * Power Go Line
 * http://cixtor.com/
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

import "flag"
import "os"

func main() {
	flag.Parse()

	var pogol PowerGoLine
	var config Configuration
	var status string = flag.Arg(0)
	var pcolor PowerColor = config.Values()

	pogol.Username(pcolor)
	pogol.Hostname()
	pogol.WorkingDirectory(pcolor, status)
	pogol.RootSymbol(pcolor, status)

	os.Exit(0)
}
