declare module '@wangeditor/editor-for-vue' {
  import { DefineComponent } from 'vue'
  
  const Editor: DefineComponent<{}, {}, any>
  const Toolbar: DefineComponent<{}, {}, any>
  
  export { Editor, Toolbar }
}