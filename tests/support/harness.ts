import {mkdirSync,writeFileSync,readFileSync} from 'node:fs'
import {resolve} from 'node:path'
import type {Page} from '@playwright/test'
import {compileSFC,convert} from './compile'
import {disclosure,checkbox,controlInput} from './inputs'
export type Mode='vapor'|'interop'
const root=resolve('.test-output/browser')
export async function buildHarnesses():Promise<void>{
 mkdirSync(root,{recursive:true})
 const sources={Disclosure:readFileSync(resolve("internal/testinput/accordion-v2.ts"),"utf8"),NamedScope:disclosure.replace('scope: { open: boolean }','scope: { name: boolean }').replace('slots.content({ open })','slots.content({ name: open })'),Binary:checkbox,Text:controlInput('text'),NumberControl:controlInput('number'),Range:controlInput('range')}
 for(const [name,source]of Object.entries(sources))writeFileSync(resolve(root,`component-${name}.ts`),compileSFC(convert(source).source,`${name}.vue`))
 const cases:Record<string,{script:string;template:string}>={
  disclosure:{script:`import Disclosure from './component-Disclosure'; const open=ref(false); const second=ref(false); const toggles=ref(0);`,template:`<Disclosure v-model:open="open" name="group" @toggle="toggles++"><template #summary>First</template><template #content="{open}"><span data-slot-state>{{open?'open':'closed'}}</span></template></Disclosure><Disclosure v-model:open="second" name="group"><template #summary>Second</template><template #content="{open}"><span data-slot-state>{{open?'open':'closed'}}</span></template></Disclosure><output data-model>{{String(open)}}</output><output data-toggles>{{toggles}}</output>`},
  uncontrolled:{script:`import Disclosure from './component-Disclosure'; const open=ref(false); const name=ref('before');`,template:`<Disclosure :open="open" :name="name"><template #summary>Heading</template><template #content="{open}"><span data-slot-state>{{open?'open':'closed'}}</span></template></Disclosure><button data-action="name" @click="name='after'">Name</button><button data-action="open" @click="open=!open">Open</button>`},
  checkbox:{script:`import Binary from './component-Binary'; const checked=ref(false); const events=ref<string[]>([]); const form=ref<HTMLFormElement|null>(null); function event(e:Event){events.value.push(e.type+':'+(e.currentTarget as Element).tagName)}`,template:`<form ref="form"><Binary v-model:checked="checked" @input="event" @change="event"><template #label>Check</template></Binary></form><button data-action="reset" @click="form?.reset()">Reset</button><output data-model>{{String(checked)}}</output><output data-events>{{events.join(',')}}</output>`},
 }
 const textScript=`import Text from './component-Text'; const value=ref('initial'); const form=ref<HTMLFormElement|null>(null); const cancel=ref(false); const show=ref(true); function dispose(){show.value=false;form.value?.reset()}`
 const textTemplate=`<form ref="form" @reset="(e:Event)=>{if(cancel)e.preventDefault()}"><Text v-if="show" v-model:value="value"><template #label>Text</template><template #content="scope"><span data-slot-state>{{scope.value}}</span></template></Text></form><button data-action="reset" @click="form?.reset()">Reset</button><button data-action="parent" @click="value='parent'">Parent</button><button data-action="cancel" @click="cancel=true">Cancel</button><button data-action="dispose" @click="dispose">Dispose</button><output data-model>{{value}}</output>`
 cases.text={script:textScript,template:textTemplate}
 cases['named-scope']={script:`import NamedScope from './component-NamedScope';`,template:`<NamedScope><template #summary>Heading</template><template #content="scope"><span data-slot-state>{{scope.name?'open':'closed'}}</span></template></NamedScope>`}
 cases.number={script:`import NumberControl from './component-NumberControl';const value=ref(12);`,template:`<NumberControl v-model:value="value"><template #label>Number</template><template #content="scope"><span data-slot-state>{{Number.isNaN(scope.value)?'NaN':String(scope.value)}}</span></template></NumberControl><output data-model>{{Number.isNaN(value)?'NaN':String(value)}}</output>`}
 cases.range={script:`import Range from './component-Range';const value=ref(999);const max=ref(20);`,template:`<Range v-model:value="value" :min="10" :max="max" :step="2"><template #label>Range</template></Range><button data-action="max" @click="max=14">Max</button><output data-model>{{String(value)}}</output>`}
 cases.reassociate={script:`import Text from './component-Text';const value=ref('initial');const form=ref('first');const second=ref<HTMLFormElement|null>(null);`,template:`<form id="first"></form><form id="second" ref="second"></form><Text v-model:value="value" :form="form"><template #label>Text</template></Text><button data-action="associate" @click="form='second'">Associate</button><button data-action="reset-second" @click="second?.reset()">Reset second</button><output data-model>{{value}}</output>`}
 for(const [name,c]of Object.entries(cases)){
  const app=`<script setup lang="ts" vapor>import {ref} from 'vue';${c.script}</script><template><main>${c.template}</main></template>`
  writeFileSync(resolve(root,`app-${name}.ts`),compileSFC(app,`${name}.vue`))
  for(const mode of ['vapor','interop'] as const){
   const entry=`import {createVaporApp,createApp,vaporInteropPlugin} from 'vue';import App from './app-${name}';${mode==='vapor'?'createVaporApp(App)':'createApp(App).use(vaporInteropPlugin)'}.mount('#app');window.__htmlUiReady=true;`
   const path=resolve(root,`${name}-${mode}.entry.ts`);writeFileSync(path,entry)
   const result=await Bun.build({entrypoints:[path],target:'browser',outdir:root,naming:`${name}-${mode}.js`,define:{__DEV__:'true',__VUE_OPTIONS_API__:'false',__VUE_PROD_DEVTOOLS__:'false',__VUE_PROD_HYDRATION_MISMATCH_DETAILS__:'false'}})
   if(!result.success)throw new Error(result.logs.map(String).join('\n'))
   writeFileSync(resolve(root,`${name}-${mode}.html`),`<!doctype html><html><body><div id="app"></div><script type="module" src="/${name}-${mode}.js"></script></body></html>`)
  }
 }
}
export async function openCase(page:Page,name:string,mode:Mode='vapor'):Promise<void>{
 const crash=page.waitForEvent('pageerror').then(err=>{throw err})
 await Promise.race([
  page.goto(`http://127.0.0.1:4175/${name}-${mode}.html`).then(response=>{
   if(!response?.ok())throw new Error(`Harness ${name}-${mode}: HTTP ${response?.status()}`)
   return page.waitForFunction(()=>Boolean((window as any).__htmlUiReady))
  }),
  crash,
 ])
}
