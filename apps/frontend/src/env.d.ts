/// <reference types="vite/client" />

declare module '*.vue' {
  import { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module '*.yaml' {
  const content: string
  export default content
}

declare module '*.yml' {
  const content: string
  export default content
}