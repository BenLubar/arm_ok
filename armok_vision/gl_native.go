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
	window, err = glfw.CreateWindow(Width, Height, "arm_ok", nil, nil)
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
	glfw.PollEvents()
	return window.ShouldClose()
}

const (
	KeyA = glfw.KeyA
	KeyB = glfw.KeyB
	KeyC = glfw.KeyC
	KeyD = glfw.KeyD
	KeyE = glfw.KeyE
	KeyF = glfw.KeyF
	KeyG = glfw.KeyG
	KeyH = glfw.KeyH
	KeyI = glfw.KeyI
	KeyJ = glfw.KeyJ
	KeyK = glfw.KeyK
	KeyL = glfw.KeyL
	KeyM = glfw.KeyM
	KeyN = glfw.KeyN
	KeyO = glfw.KeyO
	KeyP = glfw.KeyP
	KeyQ = glfw.KeyQ
	KeyR = glfw.KeyR
	KeyS = glfw.KeyS
	KeyT = glfw.KeyT
	KeyU = glfw.KeyU
	KeyV = glfw.KeyV
	KeyW = glfw.KeyW
	KeyX = glfw.KeyX
	KeyY = glfw.KeyY
	KeyZ = glfw.KeyZ

	KeyUp       = glfw.KeyUp
	KeyDown     = glfw.KeyDown
	KeyLeft     = glfw.KeyLeft
	KeyRight    = glfw.KeyRight
	KeyPageUp   = glfw.KeyPageUp
	KeyPageDown = glfw.KeyPageDown
)

func IsKeyPressed(key glfw.Key, repeat bool) bool {
	switch window.GetKey(key) {
	case glfw.Press:
		return true
	case glfw.Repeat:
		return repeat
	default:
		return false
	}
}

var (
	Program      uint32
	SSAO         uint32
	FrameBuffer1 uint32
	DepthBuffer  uint32
	NormalBuffer uint32
	FrameBuffer2 uint32
	SSAOBuffer   uint32
	UniPass      int32

	UniProjection int32
	UniCamera     int32
	UniModel      int32
	UniInverse    int32

	UniAmbient     int32
	UniDirection   int32
	UniDirectional int32

	AttrScreen uint32
	AttrVert   uint32
	AttrColor  uint32
	AttrNormal uint32
)

func MakeShader(vertex, fragment string) uint32 {
	vs := gl.CreateShader(gl.VERTEX_SHADER)
	defer gl.DeleteShader(vs)
	vsource := gl.Str(vertex + "\x00")
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
	fsource := gl.Str(fragment + "\x00")
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

	program := gl.CreateProgram()
	gl.AttachShader(program, vs)
	gl.AttachShader(program, fs)
	gl.LinkProgram(program)

	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var l int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &l)

		b := make([]byte, l+1)
		gl.GetProgramInfoLog(program, l, nil, &b[0])

		panic("linking failed: " + string(b[:l]))
	}

	return program
}

func SetupGL() {
	SSAO = MakeShader(VertexSSAO, FragmentSSAO)

	gl.UseProgram(SSAO)

	gl.Uniform1i(gl.GetUniformLocation(SSAO, gl.Str("normal\x00")), 0)
	gl.Uniform1i(gl.GetUniformLocation(SSAO, gl.Str("depth\x00")), 1)

	gl.Uniform3fv(gl.GetUniformLocation(SSAO, gl.Str("kernel\x00")), int32(len(Kernel))/3, &Kernel[0])

	AttrScreen = uint32(gl.GetAttribLocation(SSAO, gl.Str("screen\x00")))
	gl.EnableVertexAttribArray(AttrScreen)

	Program = MakeShader(VertexShader, FragmentShader)

	gl.UseProgram(Program)

	UniPass = gl.GetUniformLocation(Program, gl.Str("pass\x00"))
	UniProjection = gl.GetUniformLocation(Program, gl.Str("projection\x00"))
	UniCamera = gl.GetUniformLocation(Program, gl.Str("camera\x00"))
	UniModel = gl.GetUniformLocation(Program, gl.Str("model\x00"))
	UniInverse = gl.GetUniformLocation(Program, gl.Str("inverse\x00"))

	UniAmbient = gl.GetUniformLocation(Program, gl.Str("ambient\x00"))
	UniDirection = gl.GetUniformLocation(Program, gl.Str("direction\x00"))
	UniDirectional = gl.GetUniformLocation(Program, gl.Str("directional\x00"))

	gl.Uniform1i(gl.GetUniformLocation(Program, gl.Str("ssao\x00")), 0)
	gl.Uniform2f(gl.GetUniformLocation(Program, gl.Str("screen_size\x00")), float32(Width), float32(Height))

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

	gl.GenFramebuffers(1, &FrameBuffer1)
	gl.GenTextures(1, &DepthBuffer)
	gl.GenTextures(1, &NormalBuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, FrameBuffer1)

	gl.BindTexture(gl.TEXTURE_2D, DepthBuffer)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT, int32(Width2), int32(Height2), 0, gl.DEPTH_COMPONENT, gl.UNSIGNED_SHORT, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, DepthBuffer, 0)

	gl.BindTexture(gl.TEXTURE_2D, NormalBuffer)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(Width2), int32(Height2), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, NormalBuffer, 0)

	gl.GenFramebuffers(1, &FrameBuffer2)
	gl.GenTextures(1, &SSAOBuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, FrameBuffer2)

	gl.BindTexture(gl.TEXTURE_2D, SSAOBuffer)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(Width2), int32(Height2), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, SSAOBuffer, 0)

	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	UnitBuffer = MakeBuffer(UnitData)
	NotLoadedBuffer = MakeBuffer(NotLoadedData)
	ScreenBuffer = MakeBuffer(ScreenData)
}

func MakeBuffer(data []float32) Buffer {
	var buffer uint32
	gl.GenBuffers(1, &buffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)

	gl.BufferData(gl.ARRAY_BUFFER, len(data)*float32_size, gl.Ptr(data), gl.STATIC_DRAW)

	return Buffer{
		Buffer: buffer,
		Size:   int32(len(data)),
	}
}

func PositionCamera(camera mgl32.Mat4) {
	gl.UniformMatrix4fv(UniCamera, 1, false, &camera[0])
}

var UnitBuffer, NotLoadedBuffer, ScreenBuffer Buffer
var Buffers = make(map[[3]int32]Buffer)

type Buffer struct {
	Buffer uint32
	Size   int32
}

func CleanMap() {
	dirtyLock.Lock()
	defer dirtyLock.Unlock()

	for pos, data := range Dirty {
		if old, ok := Buffers[pos]; ok {
			gl.DeleteBuffers(1, &old.Buffer)
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
	gl.Viewport(0, 0, int32(Width2), int32(Height2))

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
		gl.UniformMatrix4fv(UniInverse, 1, false, &ident[0])

		for dx := int32(-rangeX); dx <= rangeX; dx++ {
			for dy := int32(-rangeY); dy <= rangeY; dy++ {
				for dz := int32(-rangeZdown); dz <= rangeZup; dz++ {
					pos := [3]int32{
						center[0] + dx,
						center[1] + dy,
						center[2] + dz,
					}
					translate := mgl32.Translate3D(float32(pos[0])*16, float32(pos[1])*16, float32(pos[2]))
					gl.UniformMatrix4fv(UniModel, 1, false, &translate[0])
					// We don't set the inverse matrix because we are only translating.

					buffer, ok := Buffers[pos]
					if !ok {
						buffer = NotLoadedBuffer
					}
					const stride = 3 + 3 + 3
					gl.BindBuffer(gl.ARRAY_BUFFER, buffer.Buffer)
					gl.VertexAttribPointer(AttrVert, 3, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(0*float32_size))
					gl.VertexAttribPointer(AttrColor, 3, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(3*float32_size))
					gl.VertexAttribPointer(AttrNormal, 3, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(6*float32_size))

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
			gl.UniformMatrix4fv(UniModel, 1, false, &transform[0])
			transform = transform.Inv().Transpose()
			gl.UniformMatrix4fv(UniInverse, 1, false, &transform[0])

			const stride = 3 + 3 + 3
			gl.BindBuffer(gl.ARRAY_BUFFER, UnitBuffer.Buffer)
			gl.VertexAttribPointer(AttrVert, 3, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(0*float32_size))
			gl.VertexAttribPointer(AttrColor, 3, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(3*float32_size))
			gl.VertexAttribPointer(AttrNormal, 3, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(6*float32_size))

			gl.DrawArrays(gl.TRIANGLES, 0, UnitBuffer.Size/stride)
		}
	}
	drawTheThings()

	gl.UseProgram(SSAO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, FrameBuffer2)
	gl.Viewport(0, 0, int32(Width2), int32(Height2))

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, NormalBuffer)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, DepthBuffer)

	const stride = 2
	gl.BindBuffer(gl.ARRAY_BUFFER, ScreenBuffer.Buffer)
	gl.VertexAttribPointer(AttrScreen, 2, gl.FLOAT, false, stride*float32_size, gl.PtrOffset(0*float32_size))

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, ScreenBuffer.Size/stride)

	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.UseProgram(Program)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, int32(Width), int32(Height))

	gl.ClearColor(ambient[0], ambient[1], ambient[2], 1)
	gl.ClearDepth(1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.Uniform1i(UniPass, 1)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, SSAOBuffer)
	drawTheThings()
	gl.BindTexture(gl.TEXTURE_2D, 0)

	window.SwapBuffers()
}
