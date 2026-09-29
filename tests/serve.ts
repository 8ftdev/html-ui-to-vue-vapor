import {resolve} from 'node:path'
import {buildHarnesses} from './support/harness'
await buildHarnesses()
const root=resolve('.test-output/browser')
Bun.serve({hostname:'127.0.0.1',port:4175,async fetch(request){
 const path=new URL(request.url).pathname
 if(path==='/')return new Response('ready')
 if(!/^\/[a-z-]+\.(?:html|js)$/.test(path))return new Response('not found',{status:404})
 const file=Bun.file(resolve(root,path.slice(1)))
 return await file.exists()?new Response(file):new Response('not found',{status:404})
}})
