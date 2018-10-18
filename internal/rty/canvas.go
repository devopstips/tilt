package rty

import (
	"fmt"

	"github.com/windmilleng/tcell"
)

// Canvases hold content.

type Canvas interface {
	Size() (int, int)
	SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) error
	Close() (int, int)
	GetContent(x, y int) (mainc rune, combc []rune, style tcell.Style, width int)
}

func totalHeight(canvases []Canvas) int {
	total := 0
	for _, c := range canvases {
		_, h := c.Size()
		total += h
	}
	return total
}

// Implementations below
type cell struct {
	ch    rune
	style tcell.Style
}

type TempCanvas struct {
	width  int
	height int
	cells  [][]cell
	style  tcell.Style
}

type lineRange struct {
	start int
	end   int
}

func newTempCanvas(width, height int, style tcell.Style) *TempCanvas {
	c := &TempCanvas{width: width, height: height}
	if height != GROW {
		c.cells = make([][]cell, height)
		for i := 0; i < height; i++ {
			c.cells[i] = c.makeRow()
		}
	}
	return c
}

func (c *TempCanvas) Size() (int, int) {
	return c.width, c.height
}

func (c *TempCanvas) Close() (int, int) {
	if c.height == GROW {
		c.height = len(c.cells)
	}
	return c.width, c.height
}

func (c *TempCanvas) makeRow() []cell {
	row := make([]cell, c.width)
	for i := 0; i < c.width; i++ {
		row[i].style = c.style
	}
	return row
}

func (c *TempCanvas) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) error {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		panic(fmt.Errorf("cell %v,%v outside canvas %v,%v", x, y, c.width, c.height))
	}

	for y >= len(c.cells) {
		c.cells = append(c.cells, c.makeRow())
	}

	c.cells[y][x] = cell{ch: mainc, style: style}
	return nil
}

func (c *TempCanvas) GetContent(x, y int) (mainc rune, combc []rune, style tcell.Style, width int) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		panic(fmt.Errorf("cell %d, %d outside bounds %d, %d", x, y, c.width, c.height))
	}

	if y >= len(c.cells) {
		return 0, nil, tcell.StyleDefault, 1
	}

	cell := c.cells[y][x]
	return cell.ch, nil, cell.style, 1
}

type SubCanvas struct {
	del       Canvas
	startX    int
	startY    int
	width     int
	height    int
	highWater int
	style     tcell.Style
}

func newSubCanvas(del Canvas, startX int, startY int, width int, height int, style tcell.Style) *SubCanvas {
	_, delHeight := del.Size()
	if height == GROW && delHeight != GROW {
		panic(fmt.Errorf("can't create a growing subcanvas from a non-growing subcanvas"))
	}
	r := &SubCanvas{
		del:       del,
		startX:    startX,
		startY:    startY,
		width:     width,
		height:    height,
		highWater: -1,
		style:     style,
	}
	r.fill(-1)
	return r
}

func (c *SubCanvas) Size() (int, int) {
	return c.width, c.height
}

func (c *SubCanvas) Close() (int, int) {
	if c.height == GROW {
		c.height = c.highWater + 1
	}
	return c.width, c.height
}

func (c *SubCanvas) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) error {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return fmt.Errorf("coord %d,%d is outside bounds %d,%d", x, y, c.width, c.height)
	}

	if c.height == GROW && y > c.highWater {
		oldHighWater := c.highWater
		c.highWater = y
		c.fill(oldHighWater)
	}
	return c.del.SetContent(c.startX+x, c.startY+y, mainc, combc, style)
}

func (c *SubCanvas) fill(lastFilled int) {
	startY := lastFilled + 1
	maxY := c.height
	if maxY == GROW {
		maxY = c.highWater + 1
	}
	for y := startY; y < maxY; y++ {
		for x := 0; x < c.width; x++ {
			if err := c.del.SetContent(c.startX+x, c.startY+y, 0, nil, c.style); err != nil {
				panic(err)
			}
		}
	}
}

func (c *SubCanvas) GetContent(x int, y int) (rune, []rune, tcell.Style, int) {
	return c.del.GetContent(x, y)
}

type ScreenCanvas struct {
	del tcell.Screen
}

func newScreenCanvas(del tcell.Screen) *ScreenCanvas {
	return &ScreenCanvas{del: del}
}

func (c *ScreenCanvas) Size() (int, int) {
	return c.del.Size()
}

func (c *ScreenCanvas) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) error {
	c.del.SetContent(x, y, mainc, combc, style)
	return nil
}

func (c *ScreenCanvas) Close() (int, int) {
	return c.del.Size()
}

func (c *ScreenCanvas) GetContent(x, y int) (mainc rune, combc []rune, style tcell.Style, width int) {
	return c.del.GetContent(x, y)
}