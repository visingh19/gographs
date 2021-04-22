<template>
  <div class="graph-box">
    <div class="graph-box__actions">
      <div class="graph-box__actions-search">
        <input type="text" v-model="searchText" />
        <div v-on:click="callSearch" class="btn">Search</div>
      </div>
      <div v-on:click="simpleClick" class="btn">Reset Zoom!</div>
      <div v-on:click="simpleClick" class="btn">Zoom In!</div>
      <div v-on:click="simpleClick" class="btn">Zoom Out!</div>
      <div v-on:click="resetD3" class="btn">Reset Visual!</div>
      <div v-on:click="callGraphReset" class="btn">New Graph!</div>
    </div>
    
    <div class="graph-box__d3-container" id="graph-box__d3-id">
      <div v-if="graphNodes.length == 0 && !dataLoading" class="graph-box__no-data">No data right now!</div>
      <div v-if="dataLoading" class="graph-box__no-data">Data is loading!</div>
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
      dataLoading: false,
      graphNodes: [], //list of dictionaries
      graphLinks: [], //list of dictionaries
      searchText: "Search for a node...",
    }
  },
  methods: {
    simpleClick() {
      console.log("I was clicked.");
      this.getGraphData()
    },
    callSearch() {
      return axios.get("/api/search")
      .then((response) => {
        console.log(response)
      })
      .catch((error) =>{
        console.log("API error " + error);
        throw error
      })
    },
    callGraphReset() {
      this.dataLoading = true;
      this.graphNodes = [];
      this.graphLinks = [];
      return axios.get("/api/resetgraph")
      .then((response) => {
        console.log(response)
        this.getGraphData().then(() => {
          this.d3Init()
        });
      })
      .catch((error) =>{
        console.log("API error " + error);
        this.dataLoading = false;
        throw error
      })
    },
    // this function returns a promise ( returns axios.get... )
    getGraphData() {
      console.log("Getting graph data.");
      this.dataLoading = true;

      return axios.get("/api/graph")
      .then((response) => {
        console.log(response)
        this.graphNodes = response?.data?.nodes ?? []
        this.graphLinks = response?.data?.links ?? []
        this.dataLoading = false;
      })
      .catch((error) =>{
        console.log("API error " + error);
        this.dataLoading = false;
        throw error
      })

    },
    resetD3() {
      this.d3Delete();
      setTimeout(() => {this.d3Init();}, 500 );
    },
    d3Delete() {
      // Remove d3 graph.
      d3.select("#graph-box__d3-id").selectAll("svg").remove();
    },
    d3Init() {
      // SET UP D3 GRAPH.

      // could get width and height of area, todo...
      const width = 800, height = 400;
      const d3Nodes = this.graphNodes;
      const d3Links = this.graphLinks;

      var box = d3.select("#graph-box__d3-id").append("svg")
        .attr("width", "100%").attr("height", "100%")
        .attr("pointer-events", "all");

      var color = d3.scaleOrdinal(d3.schemeCategory10);

      var simulation = d3.forceSimulation()
        .force("link", d3.forceLink())
        .force("charge", d3.forceManyBody())
        .force("center", d3.forceCenter(width / 2, height / 2))
        .alphaTarget(.001)
        ;


      // SET UP LINKS

      var link = box.append("g")
          .attr("class", "links") // add links class to lines we will create
        .selectAll("line")
        .data(d3Links)
        .enter().append("line")
          .attr("stroke-width", function(d) { return Math.sqrt(d.value); })
          .attr("stroke", "#CCC")
          .attr("fill", "none");

      // SET UP NODES

      var node = box.append("g")
          .attr("class", "nodes") // add 'nodes' class to node groups we will create
        .selectAll("g")
        .data(d3Nodes)
        .enter().append("g")

      node.append("title")
        .text(function(d) { return d.id; });

      const radius = 5;
      // var circles = // commented out so build stops crying this is an unused var.
      node.append("circle")
          .attr("r", 5)
          .attr("fill", function(d) { return color(d.group); })
          .call(d3.drag()
              .on("start", dragstarted)
              .on("drag", dragged)
              .on("end", dragended));

      // var labels = // commented out so build stops crying this is an unused var.
      node.append("text")
          .text(function(d) {
            return d.id;
          })
          .attr('x', 6)
          .attr('y', 3);


      // FORCE FXNS

      simulation
        .nodes(d3Nodes)
        .on("tick", ticked);

      simulation.force("link")
        .links(d3Links);

      // UPDATE FUNCTION ( on 'tick' this gets called )
      // bounding logic from https://bl.ocks.org/puzzler10/2531c035e8d514f125c4d15433f79d74 
      // simply takes the maximum of 0 + radius ( so circle fits in at [0,0] ) and the lesser of
      // the current position or the edges of the box
      function ticked() {
        node
            // control the circles...
            .attr("transform", function(d) {
              const dx = Math.max(radius, Math.min(width - radius, d.x));
              const dy = Math.max(radius, Math.min(height - radius, d.y));
              return "translate(" + dx + "," + dy + ")";
            })
            // control the surrounding groups...
            .attr("cx", function(d) { return d.x = Math.max(radius, Math.min(width - radius, d.x)); })
            .attr("cy", function(d) { return d.y = Math.max(radius, Math.min(height - radius, d.y)); });
        link
            .attr("x1", function(d) { return d.source.x; })
            .attr("y1", function(d) { return d.source.y; })
            .attr("x2", function(d) { return d.target.x; })
            .attr("y2", function(d) { return d.target.y; });
      }

      // in d3v6, functions go from (d,i,nodes) to (event, d)
      function dragstarted(event, d) {
        if (event.active) simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
      }

      function dragged(event, d) {
        d.fx = event.x;
        d.fy = event.y;
      }

      function dragended(event, d) {
        if (event.active) simulation.alphaTarget(0);
        d.fx = null;
        d.fy = null;
      }
    },
    init() {
      // https://bl.ocks.org/heybignick/3faf257bbbbc7743bb72310d03b86ee8 - edited from.
      // START: GET DATA.
      this.getGraphData()
      .then(() => {
        this.d3Init()
      }) // end of then

      
      // end of init
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
