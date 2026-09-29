import { parse, compileScript } from '@vue/compiler-sfc'
import { execFileSync, spawnSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { resolve } from 'node:path'
const output=resolve('.test-output')
mkdirSync(output,{recursive:true})
const binary=resolve(output,'html-ui-to-vue-vapor')
execFileSync('go',['build','-o',binary,'./cmd/html-ui-to-vue-vapor'],{env:{...process.env,GOCACHE:resolve(output,'go-cache')}})
export function convert(source:string): {source:string;stderr:string} {
 const result=spawnSync(binary,[],{input:source,encoding:'utf8'})
 if(result.error)throw result.error
 if(result.status!==0)throw new Error(result.stderr)
 return {source:result.stdout,stderr:result.stderr}
}
export function compileSFC(source:string,filename:string):string {
 const {descriptor,errors}=parse(source,{filename})
 if(errors.length)throw new Error(errors.map(String).join('\n'))
 if(!descriptor.scriptSetup||!Object.hasOwn(descriptor.scriptSetup.attrs,'vapor'))throw new Error('missing vapor marker')
 return compileScript(descriptor,{id:filename,inlineTemplate:true}).content
}
