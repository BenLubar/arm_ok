// +build js

package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/webgl"
)

var gl *webgl.Context
var keys = make(map[int]bool) // not synchronized because JavaScript doesn't have threads and synchronization would require spawning goroutines on each key press.

func InitGL() error {
	ctx, err := webgl.NewContext(js.Global.Get("document").Call("querySelector", "#canvas"), webgl.DefaultAttributes())

	gl = ctx

	js.Global.Call("addEventListener", "keydown", func(e *js.Object) {
		code := e.Get("keyCode").Int()
		if _, ok := keys[code]; !ok {
			keys[code] = false
		}
	})
	js.Global.Call("addEventListener", "keyup", func(e *js.Object) {
		delete(keys, e.Get("keyCode").Int())
	})

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

const (
	KeyA = 'A'
	KeyB = 'B'
	KeyC = 'C'
	KeyD = 'D'
	KeyE = 'E'
	KeyF = 'F'
	KeyG = 'G'
	KeyH = 'H'
	KeyI = 'I'
	KeyJ = 'J'
	KeyK = 'K'
	KeyL = 'L'
	KeyM = 'M'
	KeyN = 'N'
	KeyO = 'O'
	KeyP = 'P'
	KeyQ = 'Q'
	KeyR = 'R'
	KeyS = 'S'
	KeyT = 'T'
	KeyU = 'U'
	KeyV = 'V'
	KeyW = 'W'
	KeyX = 'X'
	KeyY = 'Y'
	KeyZ = 'Z'

	// Source: https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/keyCode
	KeyUp       = 38
	KeyDown     = 40
	KeyLeft     = 37
	KeyRight    = 39
	KeyPageUp   = 33
	KeyPageDown = 34
)

func IsKeyPressed(key int, repeat bool) bool {
	if r, ok := keys[key]; !ok {
		return false
	} else if !r {
		keys[key] = true
		return true
	} else {
		return repeat
	}
}

var (
	Program      *js.Object
	SSAO         *js.Object
	FrameBuffer1 *js.Object
	DepthBuffer  *js.Object
	NormalBuffer *js.Object
	FrameBuffer2 *js.Object
	SSAOBuffer   *js.Object
	UniPass      *js.Object

	UniProjection *js.Object
	UniCamera     *js.Object
	UniModel      *js.Object
	UniInverse    *js.Object

	UniAmbient     *js.Object
	UniDirection   *js.Object
	UniDirectional *js.Object

	AttrScreen int
	AttrVert   int
	AttrColor  int
	AttrNormal int
)

func MakeShader(vertex, fragment string) *js.Object {
	vs := gl.CreateShader(gl.VERTEX_SHADER)
	defer gl.DeleteShader(vs)
	gl.ShaderSource(vs, vertex)
	gl.CompileShader(vs)
	if !gl.GetShaderParameterb(vs, gl.COMPILE_STATUS) {
		panic("vertex shader failed: " + gl.GetShaderInfoLog(vs))
	}

	fs := gl.CreateShader(gl.FRAGMENT_SHADER)
	defer gl.DeleteShader(fs)
	gl.ShaderSource(fs, fragment)
	gl.CompileShader(fs)
	if !gl.GetShaderParameterb(fs, gl.COMPILE_STATUS) {
		panic("fragment shader failed: " + gl.GetShaderInfoLog(fs))
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vs)
	gl.AttachShader(program, fs)
	gl.LinkProgram(program)

	if !gl.GetProgramParameterb(program, gl.LINK_STATUS) {
		panic("linking failed: " + gl.GetProgramInfoLog(program))
	}

	return program
}

func SetupGL() {
	gl.GetExtension("WEBGL_depth_texture")

	SSAO = MakeShader(VertexSSAO, FragmentSSAO)

	gl.UseProgram(SSAO)

	gl.Uniform1i(gl.GetUniformLocation(SSAO, "normal"), 0)
	gl.Uniform1i(gl.GetUniformLocation(SSAO, "depth"), 1)

	gl.Object.Call("uniform3fv", gl.GetUniformLocation(SSAO, "kernel"), Kernel)

	AttrScreen = gl.GetAttribLocation(SSAO, "screen")
	gl.EnableVertexAttribArray(AttrScreen)

	Program = MakeShader(VertexShader, FragmentShader)

	gl.UseProgram(Program)

	UniPass = gl.GetUniformLocation(Program, "pass")
	UniProjection = gl.GetUniformLocation(Program, "projection")
	UniCamera = gl.GetUniformLocation(Program, "camera")
	UniModel = gl.GetUniformLocation(Program, "model")
	UniInverse = gl.GetUniformLocation(Program, "inverse")

	UniAmbient = gl.GetUniformLocation(Program, "ambient")
	UniDirection = gl.GetUniformLocation(Program, "direction")
	UniDirectional = gl.GetUniformLocation(Program, "directional")

	gl.Uniform1i(gl.GetUniformLocation(Program, "ssao"), 0)
	gl.Uniform2f(gl.GetUniformLocation(Program, "screen_size"), float32(Width), float32(Height))

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

	FrameBuffer1 = gl.CreateFramebuffer()
	DepthBuffer = gl.CreateTexture()
	NormalBuffer = gl.CreateTexture()
	gl.BindFramebuffer(gl.FRAMEBUFFER, FrameBuffer1)

	gl.BindTexture(gl.TEXTURE_2D, DepthBuffer)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.Object.Call("texImage2D", gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, Width2, Height2, 0, gl.DEPTH_COMPONENT, gl.UNSIGNED_SHORT, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, DepthBuffer, 0)

	gl.BindTexture(gl.TEXTURE_2D, NormalBuffer)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.Object.Call("texImage2D", gl.TEXTURE_2D, 0, gl.RGB, Width2, Height2, 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, NormalBuffer, 0)

	FrameBuffer2 = gl.CreateFramebuffer()
	SSAOBuffer = gl.CreateTexture()
	gl.BindFramebuffer(gl.FRAMEBUFFER, FrameBuffer2)

	gl.BindTexture(gl.TEXTURE_2D, SSAOBuffer)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.Object.Call("texImage2D", gl.TEXTURE_2D, 0, gl.RGB, Width2, Height2, 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, SSAOBuffer, 0)

	gl.BindTexture(gl.TEXTURE_2D, nil)
	gl.BindFramebuffer(gl.FRAMEBUFFER, nil)

	UnitBuffer = MakeBuffer(UnitData)
	NotLoadedBuffer = MakeBuffer(NotLoadedData)
	ScreenBuffer = MakeBuffer(ScreenData)
}

func MakeBuffer(data []float32) Buffer {
	buffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)

	gl.BufferData(gl.ARRAY_BUFFER, data, gl.STATIC_DRAW)

	return Buffer{
		Buffer: buffer,
		Size:   len(data),
	}
}

func PositionCamera(camera mgl32.Mat4) {
	gl.UniformMatrix4fv(UniCamera, false, camera[:])
}

var UnitBuffer, NotLoadedBuffer, ScreenBuffer Buffer
var Buffers = make(map[[3]int32]Buffer)

type Buffer struct {
	Buffer *js.Object
	Size   int
}

func CleanMap() {
	dirtyLock.Lock()
	defer dirtyLock.Unlock()

	for pos, data := range Dirty {
		if old, ok := Buffers[pos]; ok {
			gl.DeleteBuffer(old.Buffer)
			delete(Buffers, pos)
		}
		if len(data) != 0 {
			Buffers[pos] = MakeBuffer(data)
		}
		delete(Dirty, pos)
	}
}

func Render(ambient, direction, directional mgl32.Vec3) {
	gl.UseProgram(Program)
	gl.BindFramebuffer(gl.FRAMEBUFFER, FrameBuffer1)
	gl.Viewport(0, 0, Width2, Height2)

	gl.ClearColor(0.5, 0.5, 0.5, 0)
	gl.ClearDepth(1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.Uniform1i(UniPass, 0)
	gl.Uniform3f(UniAmbient, ambient[0], ambient[1], ambient[2])
	gl.Uniform3f(UniDirection, direction[0], direction[1], direction[2])
	gl.Uniform3f(UniDirectional, directional[0], directional[1], directional[2])

	center := FindCenter()

	unitLock.Lock()
	units := Units
	unitLock.Unlock()

	drawTheThings := func() {
		ident := mgl32.Ident4()
		gl.UniformMatrix4fv(UniInverse, false, ident[:])

		for dx := int32(-rangeX); dx <= rangeX; dx++ {
			for dy := int32(-rangeY); dy <= rangeY; dy++ {
				for dz := int32(-rangeZdown); dz <= rangeZup; dz++ {
					pos := [3]int32{
						center[0] + dx,
						center[1] + dy,
						center[2] + dz,
					}
					translate := mgl32.Translate3D(float32(pos[0])*16, float32(pos[1])*16, float32(pos[2]))
					gl.UniformMatrix4fv(UniModel, false, translate[:])
					// We don't set the inverse matrix because we are only translating.

					buffer, ok := Buffers[pos]
					if !ok {
						buffer = NotLoadedBuffer
					}
					const stride = 3 + 3 + 3
					gl.BindBuffer(gl.ARRAY_BUFFER, buffer.Buffer)
					gl.VertexAttribPointer(AttrVert, 3, gl.FLOAT, false, stride*float32_size, 0*float32_size)
					gl.VertexAttribPointer(AttrColor, 3, gl.FLOAT, false, stride*float32_size, 3*float32_size)
					gl.VertexAttribPointer(AttrNormal, 3, gl.FLOAT, false, stride*float32_size, 6*float32_size)

					gl.DrawArrays(gl.TRIANGLES, 0, buffer.Size/stride)
				}
			}
		}

		for id, u := range units {
			if center[0]-rangeX > u.Pos[0]/16 ||
				center[0]+rangeX < u.Pos[0]/16 ||
				center[1]-rangeY > u.Pos[1]/16 ||
				center[1]+rangeY < u.Pos[1]/16 ||
				center[2]-rangeZdown > u.Pos[2] ||
				center[2]+rangeZup < u.Pos[2] {
				continue
			}

			transform := mgl32.Translate3D(float32(u.Pos[0])+0.5, float32(u.Pos[1])+0.5, float32(u.Pos[2])+0.5).Mul4(mgl32.HomogRotate3DZ(float32(id) / 100)).Mul4(mgl32.Translate3D(-0.5, -0.5, -0.5))
			gl.UniformMatrix4fv(UniModel, false, transform[:])
			transform = transform.Inv().Transpose()
			gl.UniformMatrix4fv(UniInverse, false, transform[:])

			const stride = 3 + 3 + 3
			gl.BindBuffer(gl.ARRAY_BUFFER, UnitBuffer.Buffer)
			gl.VertexAttribPointer(AttrVert, 3, gl.FLOAT, false, stride*float32_size, 0*float32_size)
			gl.VertexAttribPointer(AttrColor, 3, gl.FLOAT, false, stride*float32_size, 3*float32_size)
			gl.VertexAttribPointer(AttrNormal, 3, gl.FLOAT, false, stride*float32_size, 6*float32_size)

			gl.DrawArrays(gl.TRIANGLES, 0, UnitBuffer.Size/stride)
		}
	}
	drawTheThings()

	gl.UseProgram(SSAO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, FrameBuffer2)
	gl.Viewport(0, 0, Width2, Height2)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, NormalBuffer)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, DepthBuffer)

	const stride = 2
	gl.BindBuffer(gl.ARRAY_BUFFER, ScreenBuffer.Buffer)
	gl.VertexAttribPointer(AttrScreen, 2, gl.FLOAT, false, stride*float32_size, 0*float32_size)

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, ScreenBuffer.Size/stride)

	gl.BindTexture(gl.TEXTURE_2D, nil)

	gl.UseProgram(Program)
	gl.BindFramebuffer(gl.FRAMEBUFFER, nil)
	gl.Viewport(0, 0, Width, Height)

	gl.ClearColor(ambient[0], ambient[1], ambient[2], 1)
	gl.ClearDepth(1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.Uniform1i(UniPass, 1)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, SSAOBuffer)
	drawTheThings()
	gl.BindTexture(gl.TEXTURE_2D, nil)

	gl.Flush()
}
