// Copyright 2015, Timothy` Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"fmt"
	"math"
	"os"
	"runtime"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	input "github.com/tbogdala/fizzle/input/glfwinput"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

/*
  This example illustrates the bare minimum to set up an application
  using the fizzle library.

  It does the following:

    1) creates a GFLW window for rendering
    2) creates a renderer
    3) loads some shaders
    4) creates a cube and a sphere
    5) in a loop, render the cube or sphere
		6) when escape is pressed, exit the loop
		7) when spacebar is pressed toggle which shape to draw

  This example also does not use the 'example app' framework so that
  it can be as compact and illustrative of the minimal requirements
  as possible.
*/

// GLFW event handling must run on the main OS thread. If this doesn't get
// locked down, you will likely see random crashes on memory access while
// running the application after a few seconds.
//
// So on initialization of the module, lock the OS thread for this goroutine.
func init() {
	runtime.LockOSThread()
}

const (
	windowWidth       = 800
	windowHeight      = 600
	radsPerSec        = math.Pi / 4.0
	diffuseShaderPath = "./assets/forwardshaders/diffuse"
)

var (
	mainWindow *glfw.Window
	renderer   *forward.ForwardRenderer

	objects = make(map[*fizzle.Renderable]bool)
)

// main is the entry point for the application.
func main() {
	// start off by initializing the GL and GLFW libraries and creating a window.
	// the default window size we use is 800x600
	w, gfx := initGraphics("Simple Cube", windowWidth, windowHeight)
	mainWindow = w

	// set the callback functions for key input
	kbModel := input.NewKeyboardModel(mainWindow)
	kbModel.BindTrigger(glfw.KeyEscape, setShouldClose)
	kbModel.BindTrigger(glfw.KeySpace, toggleModel)
	kbModel.SetupCallbacks()

	// create a new renderer
	renderer = forward.NewForwardRenderer(gfx)
	renderer.ChangeResolution(windowWidth, windowHeight)
	defer renderer.Destroy()

	// put a light in there
	light := renderer.NewLight()
	//light.Position = mgl.Vec3{-10.0, 5.0, 10}
	light.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	light.Direction = mgl.Vec3{1.0, -0.5, -1.0}
	light.DiffuseIntensity = 0.70
	light.SpecularIntensity = 0.10
	light.AmbientIntensity = 0.20
	light.Attenuation = 1.0
	renderer.ActiveLights[0] = light

	// setup the camera to look at the cube
	camera := fizzle.NewOrbitCamera(mgl.Vec3{0, 0, 0}, math.Pi/2.0, 5.0, math.Pi/2.0)

	// set some OpenGL flags
	gfx.Enable(graphics.CULL_FACE)
	gfx.Enable(graphics.DEPTH_TEST)

	// loop until something told the mainWindow that it should close
	for !mainWindow.ShouldClose() {

		// handle any keyboard input
		kbModel.CheckKeyPresses()

		// clear the screen
		width, height := renderer.GetResolution()
		gfx.Viewport(0, 0, int32(width), int32(height))
		gfx.ClearColor(0.25, 0.25, 0.25, 1.0)
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		// make the projection and view matrixes
		perspective := mgl.Perspective(mgl.DegToRad(60.0), float32(width)/float32(height), 1.0, 100.0)
		view := camera.GetViewMatrix()

		// draw the cube or the sphere
		for object, _ := range objects {
			renderer.DrawRenderable(object, nil, perspective, view, camera)
		}

		// draw the screen
		mainWindow.SwapBuffers()

		// advise GLFW to poll for input. without this the window appears to hang.
		glfw.PollEvents()
	}
}

// initGraphics creates an OpenGL window and initializes the required graphics libraries.
// It will either succeed or panic.
func initGraphics(title string, w int, h int) (*glfw.Window, graphics.GraphicsProvider) {
	// GLFW must be initialized before it's called
	err := glfw.Init()
	if err != nil {
		panic("Can't init glfw! " + err.Error())
	}

	// request a OpenGL 3.3 core context
	glfw.WindowHint(glfw.Samples, 0)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// do the actual window creation
	mainWindow, err = glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		panic("Failed to create the main window! " + err.Error())
	}
	mainWindow.SetSizeCallback(onWindowResize)
	mainWindow.MakeContextCurrent()

	// disable v-sync for max draw rate
	glfw.SwapInterval(0)

	// initialize OpenGL
	gfx, err := opengl.InitOpenGL()
	if err != nil {
		panic("Failed to initialize OpenGL! " + err.Error())
	}
	fizzle.SetGraphics(gfx)

	return mainWindow, gfx
}

// setShouldClose should be called to close the window and kill the app.
func setShouldClose() {
	mainWindow.SetShouldClose(true)
}

// onWindowResize is called when the window changes size
func onWindowResize(w *glfw.Window, width int, height int) {
	renderer.ChangeResolution(int32(width), int32(height))
}

// toggleModel sets whether or not the cube or the sphere should be rendered.
func toggleModel() {
	// spacebar toggles the drawing of the cube or the sphere

	if len(objects) == 0 {
		object := fizzle.CreateCube("diffuse", -1, -1, -1, 1, 1, 1)

		// load the diffuse shader
		diffuseShader, err := fizzle.LoadShaderProgramFromFiles(diffuseShaderPath, nil)
		if err != nil {
			fmt.Printf("Failed to compile and link the diffuse shader program!\n%v", err)
			os.Exit(1)
		}
		defer diffuseShader.Destroy()

		object.Core.Shader = diffuseShader
		object.Core.DiffuseColor = mgl.Vec4{0.9, 0.05, 0.05, 1.0}
		object.Core.SpecularColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}

		objects[object] = true

	} else {
		objects = make(map[*fizzle.Renderable]bool)
	}
}
