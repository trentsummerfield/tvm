package main

import (
	"flag"
	"fmt"
	"strings"

	termbox "github.com/nsf/termbox-go"
	"github.com/trentsummerfield/tvm/java"
)

func drawString(x, y int, s string, color ...termbox.Attribute) {
	fg := termbox.ColorDefault
	bg := termbox.ColorDefault
	l := len(color)
	if l > 0 {
		fg = color[0]
	}
	if l > 1 {
		bg = color[1]
	}
	for i, c := range s {
		termbox.SetCell(x+i, y, c, fg, bg)
	}
}

func drawBox(x, y, w, h int) {
	const coldef = termbox.ColorDefault

	termbox.SetCell(x, y, '╭', coldef, coldef)
	termbox.SetCell(x+w, y, '╮', coldef, coldef)
	termbox.SetCell(x, y+h, '╰', coldef, coldef)
	termbox.SetCell(x+w, y+h, '╯', coldef, coldef)
	for i := x + 1; i < x+w; i++ {
		termbox.SetCell(i, y, '─', coldef, coldef)
		termbox.SetCell(i, y+h, '─', coldef, coldef)
	}
	for i := y + 1; i < y+h; i++ {
		termbox.SetCell(x, i, '│', coldef, coldef)
		termbox.SetCell(x+w, i, '│', coldef, coldef)
	}
	for i := x + 1; i < x+w-1; i++ {
		for j := y + 1; j < y+h-1; j++ {
			termbox.SetCell(i, j, ' ', coldef, coldef)
		}
	}
}

func drawListBox(ui *ui, r *rect, title string, elements []string) {
	drawBox(r.x, r.y, r.w, r.h)
	tw := len(title)
	tx := r.x + (r.w / 2) - (tw / 2)
	drawString(tx, r.y, title)

	n := len(elements)
	visible_items := r.h - 2
	if n > visible_items {
		drawString(r.x+r.w-1, r.y+1, "^")
		drawString(r.x+r.w-1, r.y+r.h-1, "v")
	}
	for i, s := range elements {
		if i < r.h-1 {
			drawString(r.x+1, r.y+1+i, s)
		}
	}
}

func drawFrame(x, y int, frame *java.Frame) int {
	offset := 0
	if frame.PreviousFrame != nil {
		offset += drawFrame(x, y, frame.PreviousFrame)
		offset += 2
	}
	y += offset
	drawBox(x, y, 120, 120)
	x++
	y++
	class := "ROOT"
	method := ""
	sig := ""
	ret := "void"
	if frame.Class != nil {
		class = frame.Class.Name()
	}
	if frame.Method != nil {
		method = frame.Method.Name()
		sig = strings.Join(frame.Method.Signiture[:len(frame.Method.Signiture)-1], ", ")
		ret = frame.Method.Signiture[len(frame.Method.Signiture)-1]
	}
	drawString(x, y, ret+" "+class+"::"+method+"("+sig+")")
	y++
	if frame.PC != nil {
		drawString(x, y, "Byte Code Stream")
		for i, b := range frame.PC.RawByteCodes {
			fg := termbox.ColorDefault
			if i == frame.PC.RawByteCodeIndex {
				fg = termbox.ColorRed
			}
			drawString(x, y+i+1, fmt.Sprintf("%d", b), fg)
		}

		yoffset := 0
		xoffset := 20
		drawString(x+xoffset, y, "Parsed Code Stream")
		for i, b := range frame.PC.OpCodes {
			fg := termbox.ColorDefault
			if i == frame.PC.OpCodeIndex {
				fg = termbox.ColorRed
			}
			drawString(x+xoffset, y+i+1+yoffset, b.Name(), fg)
			yoffset += b.Width() - 1
		}

		yoffset = 0
		xoffset = 40
		drawString(x+xoffset, y, "Stack")
		stack := frame.Items
		for i := len(stack) - 1; i >= 0; i-- {
			item := stack[i]
			drawString(x+xoffset, y+1+yoffset, fmt.Sprintf("%v", item))
			yoffset++
		}
	}
	return offset
}

func redraw_all(vm java.VM, ui *ui) {
	if ui.mouse_click_end && ui.hot_box != nil {
		ui.hot_box.x = ui.mouse_x
		ui.hot_box.y = ui.mouse_y
	}

	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)

	y := 0
	drawString(0, y, "TVM: The Transparent Virtual Machine")
	y++
	frame := vm.ActiveFrame()
	if frame != nil {
		drawFrame(0, 2, frame)
	}
}

type rect struct {
	x, y, w, h int
}

type ui struct {
	loaded_classes    rect
	hot_box           *rect
	mouse_click_start bool
	mouse_click_end   bool
	mouse_x, mouse_y  int
}

func main() {
	batch := flag.Bool("batch", false, "Make TVM run the supplied code without visualisations")
	flag.Parse()
	vm := java.NewVM()
	for _, arg := range flag.Args() {
		vm.LoadClass(arg)
	}
	if *batch {
		vm.Run()
		return
	}
	vm.Start()

	err := termbox.Init()
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	ui := ui{
		loaded_classes: rect{
			x: 0,
			y: 3,
			w: 25,
			h: 5,
		},
	}

	redraw_all(vm, &ui)
	termbox.Flush()
mainloop:
	for {
		ui.mouse_click_start = false
		ui.mouse_click_end = false
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break mainloop
			case termbox.KeySpace:
				vm.Step()
			}
		case termbox.EventMouse:
			switch ev.Key {
			case termbox.MouseLeft:
				ui.mouse_click_start = true
			case termbox.MouseRelease:
				ui.mouse_click_end = true
			}
		case termbox.EventError:
			panic(ev.Err)
		}
		redraw_all(vm, &ui)
		termbox.Flush()
	}
}
