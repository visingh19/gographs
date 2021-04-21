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
import * as d3 from 'd3';


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
    init() {
      console.log("component init. d3 prep starting.")
      // could get width and height of area, todo...

      // const width = 800, height = 800;

      // const force = d3.layout.force()
          // .charge(-200).linkDistance(30).size([width, height]);

      const box = d3.select("#graph-box__d3-id").append("svg")
        .attr("width", "100%").attr("height", "100%")
        .attr("pointer-events", "all");

      const nodes = [{x: 30, y: 50},
              {x: 50, y: 80},
              {x: 90, y: 120}]

      box.selectAll("circle .nodes")
        .data(nodes)
        .enter()
        .append("svg:circle")
        .attr("class", "nodes")
        .attr("cx", function(d) { return d.x; })
        .attr("cy", function(d) { return d.y; })
        .attr("r", "10px")
        .attr("fill", "blue") 


    }
  },
  mounted(){
    this.init()
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
  @import './GraphBox.css';
</style>
