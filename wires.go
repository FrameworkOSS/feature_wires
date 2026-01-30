package wires

import (
	"fmt"
	"sort"
	"strings"

	"github.com/FrameworkOSS/event"
	"github.com/FrameworkOSS/feature_commands/handler"
	"github.com/FrameworkOSS/feature_wires/wire"
	"github.com/FrameworkOSS/portal"
)

const (
	KEY_RESPONSES = "\x01"
)

var (
	cmds    = []*handler.Command{cmdWire}
	cmdWire = handler.NewCommand().
		SetID("wire").
		SetName("list wires").
		SetAbout("Lists all known wires.").
		SetUsage("Optionally provide a format to use for output. Specifying raw WILL include wire data, be warned!").
		SetAliases("ls", "show", "display", "map").
		SetRequiresPreprocessing(true).
		SetArgument(cmdArgFormat).
		SetArgument(cmdArgExclude).
		SetArgument(cmdArgInclude).
		SetSubcommand(cmdWireCreate)
	cmdWireCreate = handler.NewCommand().
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
	cmdArgFormat = handler.NewCommandArg().
			SetID("format").
			SetName("format").
			SetAbout("The format to describe the command list using.").
			SetUsage("Available formats: pretty (default), csv").
			SetAliases("f", "form", "style", "type").
			SetType(handler.CommandArgTypeString).
			SetRequiresValue(true)
	cmdArgExclude = handler.NewCommandArg().
			SetID("exclude").
			SetName("exclude list").
			SetAbout("The entries to exclude from the listing.").
			SetUsage("Specify what should be excluded with a delimited list: (raw:0x00) , ; : |").
			SetAliases("e", "ex", "x", "excluded").
			SetType(handler.CommandArgTypeString).
			SetRequiresValue(true).
			SetRepeatable(true)
	cmdArgInclude = handler.NewCommandArg().
			SetID("include").
			SetName("include list").
			SetAbout("The entries to include in the listing.").
			SetUsage("Specify what should be included with a delimited list: (raw:0x00) , ; : |").
			SetAliases("i", "in", "included").
			SetType(handler.CommandArgTypeString).
			SetRequiresValue(true).
			SetRepeatable(true)
)

type Wires struct {
	locks     *portal.PortalMutex
	processor *handler.EventCommandHandler
	resps     []*event.Event
	wires     map[string]*wire.Wire
}

func NewWires() (w *Wires) {
	w = new(Wires)
	w.locks = portal.NewPortalMutex()
	w.processor = handler.NewEventCommandHandler()
	w.resps = make([]*event.Event, 0)
	w.wires = make(map[string]*wire.Wire)

	w.processor.GetCommandHandler().
		Handle(w.cmdWireList, cmdWire).
		Handle(w.cmdWireCreate, cmdWireCreate)

	return
}

func (w *Wires) respond(ctx, r *event.Event) {
	portal.EventClaim(ctx, r)
	w.storeResp(r)
}

func (w *Wires) cmdWireList(cmd *handler.Command, e *event.Event) error {
	output := "pretty"
	if format := cmd.GetArgument("format"); format != nil {
		f := format.GetValueStringToLower()
		switch f {
		case "pretty", "csv", "raw":
			output = f
		default:
			return fmt.Errorf("wires: invalid format for listing: %s", f)
		}
	}

	excluded, included, err := handler.CmdGetExcludedIncludedList(cmd)
	if err != nil {
		return err
	}

	keys := make([]string, 0)
	for key := range w.wires {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	r := event.NewEventResponse(w.ID(), nil)

	//Header
	switch output {
	case "csv":
		r.AddStringNext("id,size,fields(;)")
	}

	//Body
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		wire := w.wires[key]

		fields := make([]string, 0)
		for _, field := range wire.GetFields() {
			fields = append(fields, fmt.Sprintf("%d", field))
		}
		sort.Strings(fields)

		if excluded != nil {
			if handler.ListContains(excluded, key) {
				continue
			}
		} else if included != nil {
			if !handler.ListContains(included, key) {
				continue
			}
		}

		switch output {
		case "csv":
			if i > 0 {
				r.AddStringNext("\n")
			}
			r.AddStringNext(fmt.Sprintf("%s,%d,%s",
				key,
				len(wire.Bytes()),
				strings.Join(fields, ";"),
			))
		case "pretty":
			if i > 0 {
				r.AddStringNext("\n\n")
			}
			r.AddStringNext(fmt.Sprintf("Wire: %s\n- Size: %d",
				key,
				len(wire.Bytes()),
			))
			if len(fields) > 0 {
				r.AddStringNext("\nFields: " + strings.Join(fields, ", "))
			}
		case "raw":
			r.AddOffsetNext()
			r.AddDataNext(wire.Bytes())
		}
	}

	w.respond(e, r)
	return nil
}

func (w *Wires) cmdWireCreate(cmd *handler.Command, e *event.Event) error {
	fmt.Println(1)
	wires := cmd.GetArgumentsID("wire")
	fmt.Println(1)
	names := make([]string, 0)
	fmt.Println(1)
	for i := 0; i < len(wires); i++ {
		fmt.Println(2)
		list, err := handler.ListSplitFromArg(wires[i])
		if err != nil {
			return err
		}
		found := false
		for j := range list {
			found = false
			for k := range names {
				if list[j] == names[k] {
					found = true
					break
				}
			}
			if !found {
				names = append(names, list[j])
			}
		}
	}

	for i := 0; i < len(names); i++ {
		wire := wire.NewWire()
		w.wires[names[i]] = wire
	}

	w.respond(e, event.NewEventSuccess(w.ID()))
	return nil
}

func (w *Wires) storeResp(e *event.Event) {
	w.locks.LockKey(KEY_RESPONSES)
	w.resps = append(w.resps, e)
	w.locks.UnlockKey(KEY_RESPONSES)
}

func (w *Wires) readResp() (e *event.Event) {
	if len(w.resps) > 0 {
		w.locks.LockKey(KEY_RESPONSES)
		e = w.resps[0]
		w.resps = w.resps[1:]
		w.locks.UnlockKey(KEY_RESPONSES)
	}
	return
}

func (w *Wires) API() int {
	return 0
}

func (w *Wires) ID() string {
	return "wires"
}

func (w *Wires) Name() string {
	return "Wires"
}

func (w *Wires) Authors() []string {
	return []string{"JoshuaDoes"}
}

func (w *Wires) Description() string {
	return "Provides a simple interface for creating and utilizing dynamic structures."
}

func (w *Wires) Version() string {
	return "v0.0.1"
}

func (w *Wires) Open() error {
	w.respond(nil, handler.NewEventCommandAdd(w.ID(), cmds...))
	w.storeResp(event.NewEventReady(w.ID(), true))
	return nil
}

func (w *Wires) Close() (errs []error, retry bool) {
	for _, wire := range w.wires {
		wire.Close()
	}
	w.wires = nil
	return
}

func (w *Wires) Input(e *event.Event) error {
	return w.processor.Process(e)
}

func (w *Wires) Output() (*event.Event, error) {
	return w.readResp(), nil
}
