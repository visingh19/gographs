<template>
  <div class="hello-world">
    <div class="hello-world__header">{{ msg }}</div>
  </div>
</template>

<script>
export default {
  name: 'HelloWorld',
  props: {
    msg: String
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.hello-world {
  max-height: 20vh;
}

.hello-world__header {
  padding-top: 20px;
  padding-bottom: 10px;
  font-size: 30px;
  color: steelblue;
  border-bottom: 1px solid steelblue;
  margin-bottom: 10px;
}
</style>
