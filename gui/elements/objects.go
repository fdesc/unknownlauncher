package elements

import(
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func NewSquareButtonWithIcon(label *widget.Label,icon *canvas.Image,emptyButton *widget.Button,d float32) *fyne.Container {
	return container.New(
		layout.NewGridWrapLayout(fyne.NewSize(58,58)),
		container.NewStack(
			emptyButton,
			container.NewBorder(
				nil,
				nil,
				nil,
				nil,				
				container.NewPadded(
					container.NewBorder(
						container.NewCenter(
							container.New(
								layout.NewGridWrapLayout(fyne.NewSize(d,d)),
								icon,
							),
						),
						nil,
						nil,
						nil,
						container.NewCenter(label),
					),
				),
			),
		),
	)
}

func NewRectangleButtonWithIcon(headingLabel *widget.Label,contentLabel *widget.Label,icon *canvas.Image,emptyButton *widget.Button,w float32) *fyne.Container {
	return container.New(
		layout.NewGridWrapLayout(fyne.NewSize(w,58)),
		container.NewStack(
			emptyButton,
			container.NewBorder(
				nil,
				nil,
				nil,
				nil,	
				container.NewPadded(
					container.NewBorder(
						nil,
						nil,	
						container.NewCenter(
							container.NewPadded(
								container.New(
									layout.NewGridWrapLayout(fyne.NewSize(36,36)),
									icon,
								),
							),
						),
						nil,
						container.NewBorder(headingLabel,contentLabel,nil,nil,nil),
					),
				),
			),
		),
	)
}
