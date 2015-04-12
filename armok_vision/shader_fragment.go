package main

const FragmentShader = `
#version 100

precision mediump float;

varying vec3 v_color;

void main() {
	gl_FragColor = vec4(v_color - (0.01 / gl_FragCoord.w), 1.0);
}
`
