package main

import (
	"flag"
	"fmt"

	termbox "github.com/nsf/termbox-go"
	"github.com/trentsummerfield/tvm/java"
)

func drawString(x, y int, s string) {
	const coldef = termbox.ColorDefault
	for i, c := range s {
		termbox.SetCell(x+i, y, c, coldef, coldef)
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
	drawBox(0, 1, 20, 10)
	drawString(1, 2, fmt.Sprintf("%v::%v", frame.Class.Name(), frame.Method.Name()))
	/*
		method := vm.ActiveMethod()
		activeMethodStr := method.Class().Name() + "::" + method.Name() + "("
		activeMethodStr += strings.Join(method.Sig(), ", ")
		activeMethodStr += ")"
		drawString(ui, 0, y, "Executing "+activeMethodStr)
		y++
		classes := vm.LoadedClasses()
		classNames := make([]string, len(classes))
		for i, c := range classes {
			classNames[i] = c.Name()
		}
		if drawListBox(ui, &ui.loaded_classes, "Loaded Classes", classNames) {
			ui.hot_box = &ui.loaded_classes
		}

		if ui.mouse_click_start {
			drawString(ui, 0, y, "CLICK")
			y++
		}

		activeClass := method.Class()
		var lines []string
		lines = append(lines, "Name: "+activeClass.Name())
		lines = append(lines, "Constant Pool: ")
		for i, cpi := range activeClass.ConstantPoolItems {
			lines = append(lines, fmt.Sprintf("[%2d] %s", i+1, cpi.String()))
		}
		drawListBox(ui, &rect{x: 30, y: 3, w: 40, h: 40}, "Active Class", lines)

		lines = make([]string, 0)

		/*
			pc := method.ProgramCounter()
			for i, bc := range pc.Ops() {
				lines = append(lines, fmt.Sprintf("[%2d] %v", i+1, bc.String()))
			}
			drawListBox(ui, &rect{x: 71, y: 3, w: 25, h: 40}, "Operations", lines)
	*/
	/*
		lines = make([]string, 0)
		for i, bc := range method.Code.Instructions {
			lines = append(lines, fmt.Sprintf("[%2d] %v", i+1, bc))
		}
		drawListBox(ui, &rect{x: 97, y: 3, w: 25, h: 40}, "Byte Code", lines)
	*/

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
