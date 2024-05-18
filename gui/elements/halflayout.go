package elements

import (
   "fyne.io/fyne/v2"
   "fyne.io/fyne/v2/theme"
   "fyne.io/fyne/v2/layout"
)

type HalfLayout struct{}

func (M *HalfLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
   minSize := fyne.NewSize(0,0)
   padding := theme.Padding()
   for _, o := range objects {
      oMin := o.MinSize()
      minSize.Width = fyne.Max(oMin.Width, minSize.Width)
      minSize.Height += oMin.Height
      minSize.Height += padding
   }
   return minSize
}

func (M *HalfLayout) isSpacer(o fyne.CanvasObject) bool {
   if !o.Visible() {
      return false
   }

   spacer, ok := o.(layout.SpacerObject)
   if !ok {
      return false
   }

   return spacer.ExpandHorizontal()
}

func (M *HalfLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
   spacers := 0
   visibleObjects := 0
   total := float32(0)
   for _, o := range objects {
      if M.isSpacer(o) {
         spacers++
         continue
      }
      visibleObjects++
      total += o.MinSize().Width
   }
   padding := theme.Padding()
   extra := size.Width - total - (1 * float32(visibleObjects-1))
   spacerSize := float32(0)
   if spacers > 0 {
      spacerSize = extra / float32(spacers)
   }

   x, y := float32(0), float32(0)
   for _,o := range objects {
      if M.isSpacer(o) {
         y+=spacerSize
      }
      o.Move(fyne.NewPos(x,y))
      height := o.MinSize().Height
      y += padding + height
      o.Resize(fyne.NewSize(size.Width/2.43, height))
   }
}
