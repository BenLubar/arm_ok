package main

const VertexSSAO = `
#version 100

precision highp float;

attribute vec2 screen;

varying vec2 tex;

void main() {
	tex = screen;
	gl_Position = vec4(screen * 2.0 - 1.0, 0.5, 1.0);
}
`
