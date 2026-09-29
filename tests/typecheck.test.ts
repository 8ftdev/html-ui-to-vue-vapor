import {expect,test} from 'bun:test'
import {mkdirSync,writeFileSync} from 'node:fs'
import {resolve} from 'node:path'
import {spawnSync} from 'node:child_process'
import {convert} from './support/compile'
import {parse} from '@vue/compiler-sfc'
import {disclosure,checkbox,staticInput,controlInput} from './support/inputs'
const dir=resolve('.test-output/types')
mkdirSync(dir,{recursive:true})
for(const [name,input] of Object.entries({Disclosure:disclosure,LiteralDisclosure:disclosure.replace('name?: string','name?: "first" | "second"'),Binary:checkbox,Static:staticInput,Text:controlInput('text'),NumberControl:controlInput('number'),Range:controlInput('range')})){
 const source=convert(input).source
 writeFileSync(resolve(dir,`${name}.vue`),source)
 writeFileSync(resolve(dir,`${name}.ts`),`import {defineProps, withDefaults, defineSlots, defineEmits} from 'vue'\n${parse(source).descriptor.scriptSetup!.content}\nexport {}\n`)
}
const options={strict:true,target:'ES2022',module:'ESNext',moduleResolution:'Bundler',lib:['ES2022','DOM'],noEmit:true,types:[],skipLibCheck:true}
writeFileSync(resolve(dir,'native.json'),JSON.stringify({compilerOptions:options,include:['*.ts']}))
writeFileSync(resolve(dir,'vue.json'),JSON.stringify({compilerOptions:options,vueCompilerOptions:{strictTemplates:true},include:['*.vue','../v2/*.vue']}))
const consumer=`<script setup lang="ts">
import {ref} from 'vue'
import Disclosure from './Disclosure.vue'
const open=ref(false)
type SlotsOf<T> = T extends (...args: infer A) => any ? NonNullable<A[1]> extends {slots: infer S} ? S : never : never
const suppliedSlots: SlotsOf<typeof Disclosure> = {summary: () => null}
function toggle(event:ToggleEvent){ void event.newState }
</script>
<template><Disclosure v-model:open="open" @toggle="toggle"><template #summary>Heading</template><template #content="scope">{{scope.open ? 'open' : 'closed'}}</template></Disclosure></template>`
function checkVue(){return spawnSync('node',['scripts/vue-typecheck.cjs','--project',resolve(dir,'vue.json'),'--pretty','false'],{encoding:'utf8'})}
test('TypeScript 7 checks generated setup implementations',()=>{
 const result=spawnSync('node',['node_modules/typescript/bin/tsc','--project',resolve(dir,'native.json'),'--pretty','false'],{encoding:'utf8'})
 expect(result.stdout+result.stderr).toBe('')
 expect(result.status).toBe(0)
})
test('Vue checker verifies props, events, required/scoped slots, and named model',()=>{
 writeFileSync(resolve(dir,'consumer.vue'),consumer)
 const result=checkVue()
 expect(result.stdout+result.stderr).toBe('')
 expect(result.status).toBe(0)
})
test('Vue checker rejects invalid public contracts',()=>{
 for(const [name,bad] of Object.entries({
  boolean:consumer.replace('v-model:open="open"',':open="\'wrong\'"'),
  scope:consumer.replace('scope.open','scope.missing'),
  event:consumer.replace('event:ToggleEvent','event:KeyboardEvent'),
  model:consumer.replace('@toggle="toggle"','@update:open="(value: string) => {}"'),
  literal:consumer.replace("'./Disclosure.vue'","'./LiteralDisclosure.vue'").replace('@toggle="toggle"',':name="\'third\'"'),
  required:consumer.replace('{summary: () => null}', '{}'),
 })){
  writeFileSync(resolve(dir,'consumer.vue'),bad)
  const result=checkVue()
  expect(result.status).not.toBe(0)
  expect(result.stdout+result.stderr).toMatch(/consumer\.vue.*error TS\d+/)
  expect(result.stdout+result.stderr).toContain(name==='scope'?'error TS2339':name==='required'?'error TS2741':'error TS2322')
  expect(result.stdout+result.stderr).not.toMatch(/(?:Disclosure|Binary|Static)\.vue.*error/)
 }
 writeFileSync(resolve(dir,'consumer.vue'),consumer)
})

test('Vue exposes portable state and part override types as named exports',()=>{
 writeFileSync(resolve(dir,'ui-consumer.vue'),`<script setup lang="ts">
import {ui, contractVersion, type AccordionClasses} from '../v2/accordion.vue'
const version: 2 = contractVersion
const classes: AccordionClasses = {trigger:{state:{expanded:'bg-accent', hover:{mode:'replace',value:'bg-muted'}}}}
// @ts-expect-error summary has no disabled state
const invalid: AccordionClasses = {trigger:{state:{disabled:'opacity-50'}}}
// @ts-expect-error slots do not invent owned content parts
const missing: AccordionClasses = {content:{base:'p-4'}}
const node: 'summary' = ui.parts.trigger.node
void [version, classes, invalid, missing, node]
</script><template><div /></template>`)
 const result=checkVue()
 expect(result.stdout+result.stderr).toBe('')
 expect(result.status).toBe(0)
})
