<template>
  <div class="graph-box">
    <div class="graph-box__actions">
      <div class="graph-box__actions-search">
        <input type="text" v-model="searchText" />
        <div v-on:click="callSearch" class="btn">Search</div>
      </div>
      <div v-on:click="simpleClick" class="btn">Click me!</div>
      <div v-on:click="simpleClick" class="btn">Zoom In!</div>
      <div v-on:click="simpleClick" class="btn">Zoom Out!</div>
      <div v-on:click="simpleClick" class="btn">Reset Graph!</div>
    </div>
    
    <div class="graph-box__d3-container">
      <div id="graph-box__d3-id"></div>
      <div v-if="graphData.length == 0" class="graph-box__no-data">No data right now! Maybe it's loading!</div>
    </div>
  </div>
</template>

<script>
// this really should be just an imported js file for this component...
import axios from 'axios';


export default {
  name: 'GraphBox',
  props: {
  },
  data() {
    return {
      graphData: [],
      searchText: "Search for a node...",
    }
  },
  methods: {
    simpleClick() {
      console.log("I was clicked.");
    },
    callSearch() {
      axios.get("/api/search")
      .then((response) => {
        console.log(response)
      })
      .catch((error) =>{
        console.log("API error " + error);
      })
    },



  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
  @import './GraphBox.css';
</style>
