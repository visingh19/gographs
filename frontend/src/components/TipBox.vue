<template>
  <div class="tip-box" v-if="show">
    <div class="tip-box__header">Tips</div>
    <div>
      <p>Click reset visual to reset selected nodes and zoom.</p>
      <p>Scroll with scroll wheel to zoom.</p>
      <p>New Graph to generate a new graph!</p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'TipBox',
  props: {
    show: Boolean
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.tip-box {
  position: absolute;
  top: 20vh;
  left: 20vw;

  max-height: 60vh;
  max-width: 60vw;
  width: 60vw;
  overflow: hidden;

  color: steelblue;
  border-radius: 20px;
  border: 1px solid steelblue;
  background-color: black;
}

.tip-box__header {
  font-size: 20px;
  border-bottom: 1px solid steelblue;
  margin-bottom: 10px;
}

</style>
