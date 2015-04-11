// +build js

package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/webgl"
)

var gl *webgl.Context

func InitGL() error {
	ctx, err := webgl.NewContext(js.Global.Get("document").Call("querySelector", "#canvas"), webgl.DefaultAttributes())

	gl = ctx

	return err
}

func CleanupGL() {
	// no-op
}

var QuitChan = make(chan bool)

func ShouldQuit() bool {
	js.Global.Call("requestAnimationFrame", func() {
		go func() {
			QuitChan <- false
		}()
	})
	return <-QuitChan
}

func DoEvents() {
	// TODO
}

var (
	Program *js.Object

	UniProjection *js.Object
	UniCamera     *js.Object

	UniAmbient     *js.Object
	UniDirection   *js.Object
	UniDirectional *js.Object

	AttrVert   int
	AttrColor  int
	AttrNormal int
)

func SetupGL() {
	vs := gl.CreateShader(gl.VERTEX_SHADER)
	defer gl.DeleteShader(vs)
	gl.ShaderSource(vs, VertexShader)
	gl.CompileShader(vs)
	if !gl.GetShaderParameterb(vs, gl.COMPILE_STATUS) {
		panic("vertex shader failed: " + gl.GetShaderInfoLog(vs))
	}

	fs := gl.CreateShader(gl.FRAGMENT_SHADER)
	defer gl.DeleteShader(fs)
	gl.ShaderSource(fs, FragmentShader)
	gl.CompileShader(fs)
	if !gl.GetShaderParameterb(fs, gl.COMPILE_STATUS) {
		panic("fragment shader failed: " + gl.GetShaderInfoLog(fs))
	}

	Program = gl.CreateProgram()
	gl.AttachShader(Program, vs)
	gl.AttachShader(Program, fs)
	gl.LinkProgram(Program)

	if !gl.GetProgramParameterb(Program, gl.LINK_STATUS) {
		panic("linking failed: " + gl.GetProgramInfoLog(Program))
	}

	gl.UseProgram(Program)

	UniProjection = gl.GetUniformLocation(Program, "projection")
	UniCamera = gl.GetUniformLocation(Program, "camera")

	UniAmbient = gl.GetUniformLocation(Program, "ambient")
	UniDirection = gl.GetUniformLocation(Program, "direction")
	UniDirectional = gl.GetUniformLocation(Program, "directional")

	AttrVert = gl.GetAttribLocation(Program, "vert")
	gl.EnableVertexAttribArray(AttrVert)
	AttrColor = gl.GetAttribLocation(Program, "color")
	gl.EnableVertexAttribArray(AttrColor)
	AttrNormal = gl.GetAttribLocation(Program, "normal")
	gl.EnableVertexAttribArray(AttrNormal)

	gl.UniformMatrix4fv(UniProjection, false, Perspective[:])

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.FrontFace(gl.CW)
}

func PositionCamera(camera mgl32.Mat4) {
	gl.UniformMatrix4fv(UniCamera, false, camera[:])
}

var Buffers = make(map[[3]int32]Buffer)

type Buffer struct {
	Buffer *js.Object
	Size   int
}

func CleanMap() {
	for _, pos := range Dirty {
		if old, ok := Buffers[pos]; ok {
			gl.DeleteBuffer(old.Buffer)
			delete(Buffers, pos)
		}
		if block, ok := Map[pos]; ok {
			data := block.Generate(pos)
			if len(data) == 0 {
				continue
			}

			buffer := gl.CreateBuffer()
			gl.BindBuffer(gl.ARRAY_BUFFER, buffer)

			gl.BufferData(gl.ARRAY_BUFFER, data, gl.STATIC_DRAW)

			Buffers[pos] = Buffer{
				Buffer: buffer,
				Size:   len(data),
			}
		}
	}

	Dirty = Dirty[:0]
}

func Render(ambient, direction, directional mgl32.Vec3) {
	gl.ClearColor(ambient[0], ambient[1], ambient[2], 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.Uniform3f(UniAmbient, ambient[0], ambient[1], ambient[2])
	gl.Uniform3f(UniDirection, direction[0], direction[1], direction[2])
	gl.Uniform3f(UniDirectional, directional[0], directional[1], directional[2])

	for dx := int32(-rangeX); dx <= rangeX; dx++ {
		for dy := int32(-rangeY); dy <= rangeY; dy++ {
			for dz := int32(-rangeZdown); dz <= rangeZup; dz++ {
				pos := [3]int32{
					(ViewInfo.GetViewPosX()+(ViewInfo.GetViewSizeX()/2))/16 + dx,
					(ViewInfo.GetViewPosY()+(ViewInfo.GetViewSizeY()/2))/16 + dy,
					ViewInfo.GetViewPosZ() + dz,
				}
				if buffer, ok := Buffers[pos]; ok {
					const stride = 3 + 3 + 3
					gl.BindBuffer(gl.ARRAY_BUFFER, buffer.Buffer)
					gl.VertexAttribPointer(AttrVert, 3, gl.FLOAT, false, stride*float32_size, 0*float32_size)
					gl.VertexAttribPointer(AttrColor, 3, gl.FLOAT, false, stride*float32_size, 3*float32_size)
					gl.VertexAttribPointer(AttrNormal, 3, gl.FLOAT, false, stride*float32_size, 6*float32_size)

					gl.DrawArrays(gl.TRIANGLES, 0, buffer.Size/stride)
				}
			}
		}
	}
	gl.Flush()
}
