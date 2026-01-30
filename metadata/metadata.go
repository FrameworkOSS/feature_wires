package metadata

import "github.com/FrameworkOSS/feature_commands/handler"

const (
	KEY_RESPONSES = "\x01"

	API         = 0
	ID          = "wires"
	Name        = "Wires"
	Authors     = "JoshuaDoes"
	Description = "Provides a simple interface for creating and utilizing dynamic structures."
	Version     = "v0.0.1"
)

var (
	Commands = []*handler.Command{CmdWire}
	CmdWire  = handler.NewCommand().
			SetID("wire").
			SetName("list wires").
			SetAbout("Lists all known wires.").
			SetUsage("Optionally provide a format to use for output. Specifying raw WILL include wire data, be warned!").
			SetAliases("ls", "show", "display", "map").
			SetRequiresPreprocessing(true).
			SetArgument(CmdArgFormat).
			SetArgument(CmdArgExclude).
			SetArgument(CmdArgInclude).
			SetSubcommand(CmdWireCreate)
	CmdWireCreate = handler.NewCommand().
			SetID("create").
			SetName("create wires").
			SetAbout("Creates one or more wires.").
			SetUsage("Provide one or more names of wires to be made.").
			SetAliases("c", "make", "new", "alloc", "allocate").
			SetRequiresPreprocessing(true).
			SetRequiresArguments(true).
			SetArgument(handler.NewCommandArg().
				SetID("wire").
				SetName("wire").
				SetAbout("The name of the wire to be made.").
				SetUsage("Provide the name of the wire to make.").
				SetAliases("w", "name", "n").
				SetType(handler.CommandArgTypeString).
				SetRepeatable(true).
				SetRequired(true).
				SetRequiresValue(true),
		)
)

/* --- SHARED COMMAND ARGUMENTS --- */
var (
	CmdArgFormat = handler.NewCommandArg().
			SetID("format").
			SetName("format").
			SetAbout("The format to describe the command list using.").
			SetUsage("Available formats: pretty (default), csv").
			SetAliases("f", "form", "style", "type").
			SetType(handler.CommandArgTypeString).
			SetRequiresValue(true)
	CmdArgExclude = handler.NewCommandArg().
			SetID("exclude").
			SetName("exclude list").
			SetAbout("The entries to exclude from the listing.").
			SetUsage("Specify what should be excluded with a delimited list: (raw:0x00) , ; : |").
			SetAliases("e", "ex", "x", "excluded").
			SetType(handler.CommandArgTypeString).
			SetRequiresValue(true).
			SetRepeatable(true)
	CmdArgInclude = handler.NewCommandArg().
			SetID("include").
			SetName("include list").
			SetAbout("The entries to include in the listing.").
			SetUsage("Specify what should be included with a delimited list: (raw:0x00) , ; : |").
			SetAliases("i", "in", "included").
			SetType(handler.CommandArgTypeString).
			SetRequiresValue(true).
			SetRepeatable(true)
)
