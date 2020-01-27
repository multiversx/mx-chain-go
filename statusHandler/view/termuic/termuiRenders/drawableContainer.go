package termuiRenders

import (
	"github.com/gizak/termui/v3"
)

const topHeight = 22

// DrawableContainer defines a container of drawable object with position and dimensions
type DrawableContainer struct {
	topLeft   termui.Drawable
	topRight  termui.Drawable
	bottom    termui.Drawable
	minHeight int
	minWidth  int
	maxWidth  int
	maxHeight int
}

//NewDrawableContainer method is used to return a new NewDrawableContainer structure
func NewDrawableContainer() *DrawableContainer {
	dc := DrawableContainer{}
	return &dc
}

// TopLeft gets the topLeft drawable
func (imh *DrawableContainer) TopLeft() termui.Drawable {
	return imh.topLeft
}

// SetTopLeft sets the topLeft drawable
func (imh *DrawableContainer) SetTopLeft(drawable termui.Drawable) {
	imh.topLeft = drawable
}

// TopRight sets the topRight drawable
func (imh *DrawableContainer) TopRight() termui.Drawable {
	return imh.topRight
}

// SetTopRight sets the topRight drawable
func (imh *DrawableContainer) SetTopRight(drawable termui.Drawable) {
	imh.topRight = drawable
}

// Bottom sets the bottom drawable
func (imh *DrawableContainer) Bottom() termui.Drawable {
	return imh.bottom
}

// SetBottom sets the bottom drawable
func (imh *DrawableContainer) SetBottom(drawable termui.Drawable) {
	imh.bottom = drawable
}

// Items returns the containing items list
func (imh *DrawableContainer) Items() []termui.Drawable {
	items := make([]termui.Drawable, 0)
	items = append(items, imh.topLeft)
	items = append(items, imh.topRight)
	items = append(items, imh.bottom)
	return items
}

// SetRectangle sets the rectangle of this drawable object
func (imh *DrawableContainer) SetRectangle(startWidth, startHeight, termWidth, termHeight int) {
	imh.maxWidth = termWidth
	imh.maxHeight = termHeight
	imh.minWidth = startWidth
	imh.minHeight = startHeight

	if imh.topLeft != nil {
		imh.topLeft.SetRect(startWidth, startHeight, imh.maxWidth/2, topHeight)
	}

	if imh.topRight != nil {
		imh.topRight.SetRect(imh.maxWidth/2, startHeight, imh.maxWidth, topHeight)
	}

	if imh.bottom != nil {
		imh.bottom.SetRect(startWidth, topHeight, imh.maxWidth, imh.maxHeight)
	}

}
