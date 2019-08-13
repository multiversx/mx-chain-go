package termuiRenders

import (
	"github.com/gizak/termui/v3"
)

type DrawableContainer struct {
	topLeft   termui.Drawable
	topRight  termui.Drawable
	bottom    termui.Drawable
	minHeight int
	minWidth  int
	maxWidth  int
	maxHeight int
	items     []termui.Drawable
}

//NewDrawableContainer method is used to return a new NewDrawableContainer structure
func NewDrawableContainer() *DrawableContainer {
	dc := DrawableContainer{}
	return &dc
}

// SetTopLeft sets the topLeft drawable
func (imh *DrawableContainer) SetTopLeft(drawable termui.Drawable) {
	imh.topLeft = drawable
}

// SetTopRight sets the topRight drawable
func (imh *DrawableContainer) SetTopRight(drawable termui.Drawable) {
	imh.topRight = drawable
}

// SetBottom sets the bottom drawable
func (imh *DrawableContainer) SetBottom(drawable termui.Drawable) {
	imh.bottom = drawable
}

// SetTopLeft sets the topLeft drawable
func (imh *DrawableContainer) GetTopLeft() termui.Drawable {
	return imh.topLeft
}

// SetTopRight sets the topRight drawable
func (imh *DrawableContainer) GetTopRight() termui.Drawable {
	return imh.topRight
}

// SetBottom sets the bottom drawable
func (imh *DrawableContainer) GetBottom() termui.Drawable {
	return imh.bottom
}

// SetBottom sets the bottom drawable
func (imh *DrawableContainer) GetItems() []termui.Drawable {
	items := make([]termui.Drawable, 0)
	items = append(items, imh.topLeft)
	items = append(items, imh.topRight)
	items = append(items, imh.bottom)
	return items
}

func (imh *DrawableContainer) SetRect(startWidth, startHeight, termWidth, termHeight int) {
	imh.maxWidth = termWidth
	imh.maxHeight = termHeight
	imh.minWidth = startWidth
	imh.minHeight = startHeight

	if imh.topLeft != nil {
		imh.topLeft.SetRect(startWidth, startHeight, imh.maxWidth/2, 20)
	}

	if imh.topRight != nil {
		imh.topRight.SetRect(imh.maxWidth/2, startHeight, imh.maxWidth, 20)
	}

	if imh.bottom != nil {
		imh.bottom.SetRect(startWidth, 20, imh.maxWidth, imh.maxHeight)
	}

}
