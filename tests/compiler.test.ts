import { expect, test } from 'bun:test'
import { convert, compileSFC } from './support/compile'
import { disclosure, checkbox, staticInput, controlInput } from './support/inputs'

for (const [name, source] of Object.entries({ disclosure, checkbox, staticInput, text:controlInput('text'), number:controlInput('number'), range:controlInput('range') })) {
 test(`${name}: supplied contract generates a Vapor component`, () => {
  const output = convert(source)
  const code = compileSFC(output.source, `${name}.vue`)
  expect(code).toContain('defineVaporComponent')
 })
}
test('arbitrary interface and factory names compile', () => {
 const source=disclosure.replaceAll('Disclosure','CustomSection').replace('function disclose','function unrelated')
 expect(compileSFC(convert(source).source,'Whatever.vue')).toContain('defineVaporComponent')
})
test('a scope field named name remains slot data', () => {
 const source=disclosure.replace('scope: { open: boolean }','scope: { name: boolean }').replace('slots.content({ open })','slots.content({ name: open })')
 const code=compileSFC(convert(source).source,'NamedScope.vue')
 expect(code).toContain('"content"')
 expect(code).toContain('name:')
})
test('escaped string defaults cannot escape the script block', () => {
 const source=disclosure.replace('name?: string;','name?: string;').replace('open: false }','open: false, name: "</script>&" }').replace('name, open','name = defaults.name, open').replace('if (name !== undefined) ','')
 const output=convert(source).source
 expect(output.match(/<\/script>/g)).toHaveLength(1)
 expect(()=>compileSFC(output,'Escaped.vue')).not.toThrow()
})

// Real producer output exercises the complete version 2 catalog through the adapter.
import { execFileSync } from 'node:child_process'
import { resolve } from 'node:path'
import { mkdirSync, writeFileSync } from 'node:fs'
const producerDir = process.env.HTML_UI_SOURCE ?? resolve('../html-ui-cli')
const producerBin = resolve('.test-output/html-ui-producer')
execFileSync('go', ['build', '-o', producerBin, './cmd/html-ui'], {cwd: producerDir})
const primitiveNames = execFileSync(producerBin, ['--list'], {encoding:'utf8'}).trim().split('\n')
const v2Dir = resolve('.test-output/v2')
mkdirSync(v2Dir, {recursive:true})
for (const name of primitiveNames) {
 test(`${name}: default producer contract preserves metadata and compiles as Vapor`, () => {
  const input = execFileSync(producerBin, [name], {encoding:'utf8'})
  expect(input).toContain('contractVersion = 2')
  const output = convert(input).source
  const compiled = compileSFC(output, `${name}.vue`)
  expect(compiled).toContain('defineVaporComponent')
  expect(output).toContain('export const ui =')
  expect(output).toContain('export const contractVersion = 2')
  expect(output).toContain(`data-ui="${name}"`)
  writeFileSync(resolve(v2Dir, `${name}.vue`), output)
 })
}
