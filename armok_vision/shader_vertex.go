package main

const VertexShader = `
#version 100

precision highp float;

uniform lowp int pass;

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform mat4 inverse;

uniform vec3 ambient;
uniform vec3 direction;
uniform vec3 directional;

attribute vec3 vert;
attribute vec3 color;
attribute vec3 normal;

varying vec3 v_color;

void main() {
	gl_Position = projection * camera * model * vec4(vert, 1.0);
	if (pass == 0) {
		v_color = (projection * camera * model * vec4(normal, 0.0)).xyz / 2.0 + 0.5;
	} else {
		v_color = color * (ambient + directional * dot((inverse * vec4(normal, 0.0)).xyz, -direction));
	}
}
`
