package main

const FragmentShader = `
#version 100

precision highp float;

uniform lowp int pass;
uniform sampler2D ssao;
uniform vec2 screen_size;

varying vec3 v_color;

void main() {
	if (pass == 0) {
		gl_FragColor = vec4(v_color, 1.0);
	} else {
		gl_FragColor = vec4(v_color - (0.01 / gl_FragCoord.w) + texture2D(ssao, gl_FragCoord.xy / screen_size).r - 0.5, 1.0);
	}
}
`
