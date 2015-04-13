package main

const FragmentSSAO = `
#version 100

precision highp float;

uniform sampler2D normal;
uniform sampler2D depth;

#define kernel_size 256
uniform vec3 kernel[kernel_size];

varying vec2 tex;

// based on http://john-chapman-graphics.blogspot.com/2013/01/ssao-tutorial.html
void main() {
	float origin = texture2D(depth, tex).r;
	vec3 norm = texture2D(normal, tex).rgb * 2.0 + 1.0;
	vec3 tangent = normalize(kernel[0] - norm * dot(kernel[0], norm));
	vec3 bitangent = cross(norm, tangent);
	mat3 tbn = mat3(tangent, bitangent, norm);
	float occlusion = 0.0;

	for (int i = 0; i < kernel_size; i++) {
		occlusion += clamp((texture2D(depth, tex + (tbn * kernel[i]).xy / 50.0).r - origin) * 1000.0, -1.0, 1.0);
	}
	occlusion /= float(kernel_size);

	gl_FragColor = vec4(vec3(occlusion / 2.0 + 0.5), 1.0);
}
`
