<template>
  <div id="app">
    <HelloWorld msg="Welcome to Go Graphs!"/>
    <GraphBox />
  </div>
</template>

<script>
import HelloWorld from './components/HelloWorld.vue'
import GraphBox from './components/GraphBox.vue'

export default {
  name: 'App',
  components: {
    HelloWorld,
    GraphBox,
  }
}
</script>

<style>
  @import "./App.css";
</style>
