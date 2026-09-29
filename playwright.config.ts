import {defineConfig} from '@playwright/test'
export default defineConfig({
 testDir:'./tests',testMatch:'**/*.pw.ts',workers:1,
 use:{headless:true,trace:'retain-on-failure'},
 projects:['chromium','firefox','webkit'].map(browserName=>({name:browserName,use:{browserName:browserName as 'chromium'|'firefox'|'webkit'}})),
 webServer:{command:'bun tests/serve.ts',url:'http://127.0.0.1:4175',reuseExistingServer:false,timeout:30000},
})
