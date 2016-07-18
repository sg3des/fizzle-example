// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	gui "github.com/tbogdala/eweygewey"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	opengl "github.com/tbogdala/fizzle/graphicsprovider/opengl"
	forward "github.com/tbogdala/fizzle/renderer/forward"
)

var (
	windowWidth       = 640
	windowHeight      = 480
	mainWindow        *glfw.Window
	gfx               graphics.GraphicsProvider
	uiman             *gui.Manager
	billboardFilepath = "assets/textures/explosion00.png"
)

const (
	fontScale    = 18
	fontFilepath = "assets/Oswald-Heavy.ttf"
	//fontFilepath = "../../examples/assets/HammersmithOne.ttf"
	fontGlyphs = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890., :[]{}\\|<>;\"'~`?/-+_=()*&^%$#@!"

	diffuseTexBumpedShaderPath = "./assets/forwardshaders/diffuse_texbumped_shadows"
	shadowmapTextureShaderPath = "./assets/forwardshaders/shadowmap_texture"
	shadowmapShaderPath        = "./assets/forwardshaders/shadowmap_generator"

	explosionPath   = "./assets/textures/explosion00.png"
	testDiffusePath = "./assets/textures/TestCube_D.png"
	testNormalsPath = "./assets/textures/TestCube_N.png"
)

// block of flags set on the command line
var (
	flagDesktopNumber int
	renderer          *forward.ForwardRenderer
)

func init() {
	runtime.LockOSThread()
	flag.IntVar(&flagDesktopNumber, "desktop", -1, "the index of the desktop to create the main window on")
}

func main() {
	// parse the command line options
	flag.Parse()

	// start off by initializing the GL and GLFW libraries and creating a window.
	mainWindow, gfx = initGraphics("go", windowWidth, windowHeight)
	renderer = forward.NewForwardRenderer(gfx)

	var blenShaderPath = "assets/forwardshaders/blend"
	// load the color shader
	blendShader, err := fizzle.LoadShaderProgramFromFiles(blenShaderPath, nil)
	if err != nil {
		panic("Failed to compile and link the color shader program! " + err.Error())
	}
	defer blendShader.Destroy()

	////////////////////////////////////////////////////////////////////////////////////////
	//
	//   initObjects
	//
	//
	//
	// load the diffuse, textured and normal mapped shader
	diffuseTexBumpedShader, err := fizzle.LoadShaderProgramFromFiles(diffuseTexBumpedShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the diffuse shader program!\n%v", err)
		os.Exit(1)
	}
	defer diffuseTexBumpedShader.Destroy()

	// load the diffuse, textured and normal mapped shader
	shadowmapTextureShader, err := fizzle.LoadShaderProgramFromFiles(shadowmapTextureShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the color texture shader program!\n%v", err)
		os.Exit(1)
	}
	defer shadowmapTextureShader.Destroy()

	// loadup the shadowmap shaders
	shadowmapShader, err := fizzle.LoadShaderProgramFromFiles(shadowmapShaderPath, nil)
	if err != nil {
		fmt.Printf("Failed to compile and link the shadowmap generator shader program!\n%v", err)
		os.Exit(1)
	}
	defer shadowmapShader.Destroy()

	// load up some textures
	textureMan := fizzle.NewTextureManager()

	explosionTex, err := textureMan.LoadTexture("explosion", explosionPath)
	if err != nil {
		fmt.Printf("Failed to load the diffuse texture at %s!\n%v", explosionPath, err)
		os.Exit(1)
	}

	diffuseTex, err := textureMan.LoadTexture("cube_diffuse", testDiffusePath)
	if err != nil {
		fmt.Printf("Failed to load the diffuse texture at %s!\n%v", testDiffusePath, err)
		os.Exit(1)
	}

	normalsTex, err := textureMan.LoadTexture("cube_diffuse", testNormalsPath)
	if err != nil {
		fmt.Printf("Failed to load the normals texture at %s!\n%v", testNormalsPath, err)
		os.Exit(1)
	}

	// var objects = make(map[*fizzle.Renderable]bool)

	smoke := fizzle.CreatePlaneXY("diffuse_texbumped", -0.5, 0.5, 0.5, -0.5)
	smoke.Scale = mgl.Vec3{10, 10, 10}
	smoke.Core.DiffuseColor = mgl.Vec4{1, 0, 0, 0.5}
	smoke.Core.Shininess = 3.0
	smoke.Core.Tex0 = explosionTex
	smoke.Core.Shader = blendShader
	smoke.Location = mgl.Vec3{0, 0, 0}
	// objects[smoke] = true

	floorPlane := fizzle.CreatePlaneXY("diffuse_texbumped", -0.5, 0.5, 0.5, -0.5)
	floorPlane.Scale = mgl.Vec3{10, 10, 10}
	floorPlane.Core.DiffuseColor = mgl.Vec4{1, 1, 1, 1}
	floorPlane.Core.SpecularColor = mgl.Vec4{0.3, 0.3, 0.3, 1.0}
	floorPlane.Core.Shininess = 3.0
	floorPlane.Core.Tex0 = diffuseTex
	floorPlane.Core.Tex1 = normalsTex
	floorPlane.Core.Shader = diffuseTexBumpedShader
	floorPlane.Location = mgl.Vec3{0, 0, -2}
	// objects[floorPlane] = true

	testCube := fizzle.CreateCube("diffuse_texbumped", -0.5, -0.5, -0.5, 0.5, 0.5, 0.5)
	testCube.Core.DiffuseColor = mgl.Vec4{1.0, 1.0, 1.0, 1.0}
	testCube.Core.SpecularColor = mgl.Vec4{0.3, 0.3, 0.3, 1.0}
	testCube.Core.Shininess = 6.0
	testCube.Core.Tex0 = diffuseTex
	testCube.Core.Tex1 = normalsTex
	testCube.Core.Shader = diffuseTexBumpedShader
	// objects[testCube] = true

	// enable shadow mapping in the renderer
	renderer.SetupShadowMapRendering()

	// add light #1
	light := renderer.NewLight()
	light.Position = mgl.Vec3{5.0, 3, 5.0}
	light.DiffuseColor = mgl.Vec4{0.9, 0.9, 0.9, 1.0}
	light.DiffuseIntensity = 30
	light.AmbientIntensity = 0.50
	light.Attenuation = 0.2
	renderer.ActiveLights[0] = light
	light.CreateShadowMap(4096, 0.5, 50.0, mgl.Vec3{-5.0, -3.0, -5.0})
	//
	//
	//////////////////////////////////////////////////////////////////////////////////////////

	camera := fizzle.NewYawPitchCamera(mgl.Vec3{0, -10, 10})
	camera.SetYawAndPitch(0.0, mgl.DegToRad(-45))
	// camera.LookAtDirect(mgl.Vec3{0, 0, 0})
	// camera.LookAt(mgl.Vec3{0, 0, 0}, 20)

	/////////////////////////////////////////////////////////////////////////////
	// loop until something told the mainWindow that it should close
	// set some OpenGL flags
	gfx.Enable(graphics.CULL_FACE)
	gfx.Enable(graphics.DEPTH_TEST)
	gfx.Enable(graphics.PROGRAM_POINT_SIZE)
	gfx.Enable(graphics.TEXTURE_2D)
	gfx.Enable(graphics.BLEND)
	gfx.BlendFunc(graphics.SRC_ALPHA, graphics.ONE_MINUS_SRC_ALPHA)

	// lastFrame := time.Now()
	for !mainWindow.ShouldClose() {

		// Shadow time!
		renderer.StartShadowMapping()
		lightCount := renderer.GetActiveLightCount()
		if lightCount >= 1 {
			for lightI := 0; lightI < lightCount; lightI++ {
				// get lights with shadow maps
				lightToCast := renderer.ActiveLights[lightI]
				if lightToCast.ShadowMap == nil {
					continue
				}

				// enable the light to cast shadows
				renderer.EnableShadowMappingLight(lightToCast)
				renderer.DrawRenderableWithShader(testCube, shadowmapShader, nil, lightToCast.ShadowMap.Projection, lightToCast.ShadowMap.View, camera)
				renderer.DrawRenderableWithShader(floorPlane, shadowmapShader, nil, lightToCast.ShadowMap.Projection, lightToCast.ShadowMap.View, camera)
				// renderer.DrawRenderableWithShader(floorPlane2, shadowmapShader, nil, lightToCast.ShadowMap.Projection, lightToCast.ShadowMap.View, camera)

			}
		}
		renderer.EndShadowMapping()

		// clear the screen
		gfx.Viewport(0, 0, int32(windowWidth), int32(windowHeight))
		gfx.ClearColor(0.1, 0.1, 0.1, 0.9)
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		// make the projection and view matrixes
		perspective := mgl.Perspective(mgl.DegToRad(60.0), float32(windowWidth)/float32(windowHeight), 0.1, 50.0)
		view := camera.GetViewMatrix()

		renderer.DrawRenderable(smoke, nil, perspective, view, camera)
		renderer.DrawRenderable(testCube, nil, perspective, view, camera)
		renderer.DrawRenderable(floorPlane, nil, perspective, view, camera)

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

	// get a list of all the monitors to use and then take the one
	// specified by the command line flag
	monitors := glfw.GetMonitors()
	if flagDesktopNumber >= len(monitors) {
		flagDesktopNumber = -1
	}
	var monitorToUse *glfw.Monitor
	if flagDesktopNumber >= 0 {
		monitorToUse = monitors[flagDesktopNumber]
	}

	// do the actual window creation
	mainWindow, err = glfw.CreateWindow(w, h, title, monitorToUse, nil)
	if err != nil {
		panic("Failed to create the main window! " + err.Error())
	}
	mainWindow.SetSizeCallback(onWindowResize)
	mainWindow.MakeContextCurrent()

	// disable v-sync for max draw rate
	glfw.SwapInterval(1)

	// initialize OpenGL
	gfx, err := opengl.InitOpenGL()
	if err != nil {
		panic("Failed to initialize OpenGL! " + err.Error())
	}
	fizzle.SetGraphics(gfx)

	return mainWindow, gfx
}

// onWindowResize is called when the window changes size
func onWindowResize(w *glfw.Window, width int, height int) {
	// uiman.AdviseResolution(int32(width), int32(height))
	// renderer.ChangeResolution(int32(width), int32(height))
}
