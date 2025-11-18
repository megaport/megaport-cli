/// <reference types="vite/client" />

// Declare module for Vue SFC imports
declare module '*.vue' {
  import type { DefineComponent } from 'vue';
  const component: DefineComponent<{}, {}, any>;
  export default component;
}

// Vue 3 Compiler Macros (these are auto-imported in <script setup>)
// The Vue compiler will transform these - they don't need to be explicitly imported
declare global {
  const defineProps: typeof import('vue')['defineProps'];
  const defineEmits: typeof import('vue')['defineEmits'];
  const defineExpose: typeof import('vue')['defineExpose'];
  const withDefaults: typeof import('vue')['withDefaults'];
}

export {};
