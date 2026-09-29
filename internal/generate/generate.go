// Package generate emits standalone Vue Vapor single-file components.
package generate

import (
	"html-ui-to-vue-vapor/internal/contract"
	"strings"
)

type Result struct {
	Source   string
	Warnings []string
}
type renderer struct {
	c                *contract.Component
	script, template strings.Builder
	fields           []stateField
}

func Generate(c *contract.Component) (Result, error) {
	if err := contract.Validate(c); err != nil {
		return Result{}, err
	}
	r := &renderer{c: c, fields: inferState(c)}
	r.writeScript()
	r.writeNode(c.Root, 1)
	result := Result{}
	if c.Behavior == "adapter-required" {
		result.Warnings = append(result.Warnings, "adapter-required: input describes native structure only; additional interaction behavior requires an adapter")
	}
	if c.Behavior == "unknown" {
		result.Warnings = append(result.Warnings, "unclassified behavior: conversion implements only the supplied native contract")
	}
	var b strings.Builder
	for _, w := range result.Warnings {
		b.WriteString("<!-- " + strings.ReplaceAll(w, "--", "—") + " -->\n")
	}
	module := r.nativeAttributeTypes()
	if module != "" {
		module = "import type {} from 'vue';\n\n" + module
	}
	if c.UI != nil {
		module += contract.UIExports(c)
	}
	if module != "" {
		b.WriteString("<script lang=\"ts\">\n" + module + "</script>\n\n")
	}
	b.WriteString("<script setup lang=\"ts\" vapor>\n")
	b.WriteString(r.script.String())
	b.WriteString("</script>\n\n<template>\n")
	b.WriteString(r.template.String())
	b.WriteString("</template>\n")
	result.Source = b.String()
	return result, nil
}
func (r *renderer) node(id string) *contract.Node {
	for i := range r.c.Nodes {
		if r.c.Nodes[i].ID == id {
			return &r.c.Nodes[i]
		}
	}
	panic("validated node missing")
}
