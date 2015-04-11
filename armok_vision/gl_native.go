// +build !js

package main

import (
	"runtime"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var window *glfw.Window

func InitGL() error {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		return err
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Samples, 8)

	var err error
	window, err = glfw.CreateWindow(800, 600, "arm_ok", nil, nil)
	if err != nil {
		return err
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		return err
	}

	return nil
}

func CleanupGL() {
	glfw.Terminate()
}

func ShouldQuit() bool {
	return window.ShouldClose()
}

func DoEvents() {
	glfw.PollEvents()
	// TODO
}

var (
	Program uint32

	UniProjection int32
	UniCamera     int32

	UniAmbient     int32
	UniDirection   int32
	UniDirectional int32

	AttrVert   uint32
	AttrColor  uint32
	AttrNormal uint32
)

func SetupGL() {
	vs := gl.CreateShader(gl.VERTEX_SHADER)
	defer gl.DeleteShader(vs)
	vsource := gl.Str(VertexShader + "\x00")
	gl.ShaderSource(vs, 1, &vsource, nil)
	gl.CompileShader(vs)
	var status int32
	gl.GetShaderiv(vs, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var l int32
		gl.GetShaderiv(vs, gl.INFO_LOG_LENGTH, &l)

		b := make([]byte, l+1)
		gl.GetShaderInfoLog(vs, l, nil, &b[0])

		panic("vertex shader failed: " + string(b[:l]))
	}

	fs := gl.CreateShader(gl.FRAGMENT_SHADER)
	defer gl.DeleteShader(fs)
	fsource := gl.Str(FragmentShader + "\x00")
	gl.ShaderSource(fs, 1, &fsource, nil)
	gl.CompileShader(fs)
	gl.GetShaderiv(fs, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var l int32
		gl.GetShaderiv(fs, gl.INFO_LOG_LENGTH, &l)

		b := make([]byte, l+1)
		gl.GetShaderInfoLog(fs, l, nil, &b[0])

		panic("fragment shader failed: " + string(b[:l]))
	}

	Program = gl.CreateProgram()
	gl.AttachShader(Program, vs)
	gl.AttachShader(Program, fs)
	gl.LinkProgram(Program)

	gl.GetProgramiv(Program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var l int32
		gl.GetProgramiv(Program, gl.INFO_LOG_LENGTH, &l)

		b := make([]byte, l+1)
		gl.GetProgramInfoLog(Program, l, nil, &b[0])

		panic("linking failed: " + string(b[:l]))
	}

	gl.UseProgram(Program)

	UniProjection = gl.GetUniformLocation(Program, gl.Str("projection\x00"))
	UniCamera = gl.GetUniformLocation(Program, gl.Str("camera\x00"))

	UniAmbient = gl.GetUniformLocation(Program, gl.Str("ambient\x00"))
	UniDirection = gl.GetUniformLocation(Program, gl.Str("direction\x00"))
	UniDirectional = gl.GetUniformLocation(Program, gl.Str("directional\x00"))

	AttrVert = uint32(gl.GetAttribLocation(Program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(AttrVert)
	AttrColor = uint32(gl.GetAttribLocation(Program, gl.Str("color\x00")))
	gl.EnableVertexAttribArray(AttrColor)
	AttrNormal = uint32(gl.GetAttribLocation(Program, gl.Str("normal\x00")))
	gl.EnableVertexAttribArray(AttrNormal)

	gl.UniformMatrix4fv(UniProjection, 1, false, &Perspective[0])

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.FrontFace(gl.CW)
}

func PositionCamera(camera mgl32.Mat4) {
	gl.UniformMatrix4fv(UniCamera, 1, false, &camera[0])
}

var Buffers = make(map[[3]int32]Buffer)

type Buffer struct {
	Buffer uint32
	Size   int32
}

func CleanMap() {
	for _, pos := range Dirty {
		if old, ok := Buffers[pos]; ok {
			gl.DeleteBuffers(1, &old.Buffer)
			delete(Buffers, pos)
		}
		if block, ok := Map[pos]; ok {
			data := block.Generate(pos)
			if len(data) == 0 {
				continue
			}

			var buffer uint32
			gl.GenBuffers(1, &buffer)
			gl.BindBuffer(gl.ARRAY_BUFFER, buffer)

			gl.BufferData(gl.ARRAY_BUFFER, len(data)*float32_size, gl.Ptr(data), gl.STATIC_DRAW)

			Buffers[pos] = Buffer{
				Buffer: buffer,
				Size:   int32(len(data)),
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
					gl.VertexAttribPointer(AttrVert, 3, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(0*float32_size))
					gl.VertexAttribPointer(AttrColor, 3, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(3*float32_size))
					gl.VertexAttribPointer(AttrNormal, 3, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(6*float32_size))

					gl.DrawArrays(gl.TRIANGLES, 0, buffer.Size/stride)
				}
			}
		}
	}
	window.SwapBuffers()
}
