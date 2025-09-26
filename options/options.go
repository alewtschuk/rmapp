package options

// Holds all command line related options
type Options struct {
	Verbosity bool // is verbose flag set
	Mode      bool // sets mode between trash and delete
	Peek      bool // sets user peeking files to true
	Logical   bool // sets whether the user wants logical or native disk usage size
	Size      bool // sets if the user just wants to view application size
}
