// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/webgl"
)

var context *webgl.Context

func InitGL() error {
	ctx, err := webgl.NewContext(js.Global.Get("document").Call("querySelector", "#canvas"), webgl.DefaultAttributes())

	context = ctx

	return err
}

func CleanupGL() {
	// no-op
}

func PositionCamera() {
	js.Global.Set("view_info", js.MakeWrapper(ViewInfo))
}

func UseBlockList() {
	js.Global.Set("block_list", js.MakeWrapper(BlockList))
}
