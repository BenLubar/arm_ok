package main

const VertexShader = `
#version 100

precision mediump float;

uniform mat4 projection;
uniform mat4 camera;

uniform vec3 ambient;
uniform vec3 direction;
uniform vec3 directional;

attribute vec3 vert;
attribute vec3 color;
attribute vec3 normal;

varying vec3 v_color;

void main() {
	gl_Position = projection * camera * vec4(vert, 1.0);
	v_color = color * (ambient + directional * dot(normal, -direction));
}
`
