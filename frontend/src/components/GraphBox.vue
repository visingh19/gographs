<template>
  <div class="graph-box">
    <div class="graph-box__actions">
      <div class="graph-box__actions-search">
        <input type="text" v-model="searchText" />
        <div v-on:click="callSearch" class="btn">Search</div>
      </div>
      <div v-on:click="resetD3" class="btn">Reset Visual!</div>
      <div v-on:click="callGraphReset" class="btn">New Graph!</div>
      <div v-on:click="showTips = !showTips" class="btn">Tips {{showTips ? "Off" : "On"}}</div>
    </div>
    
    <div class="graph-box__d3-container" id="graph-box__d3-id" ref="graphBoxRef">
      <div v-if="graphNodes.length == 0 && !dataLoading" class="graph-box__no-data">No data right now!</div>
      <div v-if="dataLoading" class="graph-box__no-data">Data is loading!</div>
    </div>

    <TipBox v-bind:show="showTips" />
  </div>
</template>


<script>
// this really should be just an imported js file for this component...
import axios from 'axios';
import * as d3 from 'd3';
import TipBox from './TipBox.vue'

export default {
  name: 'GraphBox',
  components: {
    TipBox,
  },
  props: {
  },
  data() {
    return {
      dataLoading: false,
      graphNodes: [], //list of dictionaries
      graphLinks: [], //list of dictionaries
      searchText: "Search for a node...",
      clickedNodeIdx: -1,
      d3NodeVar: null,
      d3LinkVar: null,
      showTips: false,

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
      this.d3Delete();
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
      // console.log("Getting graph data.");
      this.dataLoading = true;

      return axios.get("/api/graph")
      .then((response) => {
        // console.log(response)
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
      this.d3Init();
    },
    d3Delete() {
      // Remove d3 graph.
      d3.select("#graph-box__d3-id").selectAll("svg").remove();
    },
    d3ResetZoom() {
      // non functional. just reset the whole chart with resetD3 instead.
    },
    // this function kind of does everything...
    d3Init() {
      // SET UP D3 GRAPH.

      // could get width and height of area, todo...
      const chartGutter = 50; // sets gutter/2 px 'padding' to box's bounding functions - maybe unneeded with zoom.
      const width = this.$refs.graphBoxRef.clientWidth; 
      const height = this.$refs.graphBoxRef.clientHeight;
      const d3Nodes = this.graphNodes;
      const d3Links = this.graphLinks;

      var box = d3.select("#graph-box__d3-id").append("svg")
        .attr("id", "graph-box__d3-svg-id")
        .attr("width", "100%").attr("height", "100%")
        .attr("pointer-events", "all");

      var boxInner = box.append("g")
          .attr("class", "d3-box-inner");

      box
        .call(d3.zoom().on("zoom", function (event) {
           boxInner.attr("transform", event.transform)
        }))

      var color = d3.scaleOrdinal(d3.schemeCategory10);
      var nodeColor = d3.scaleSequential().interpolator(d3.interpolateBlues).domain([0,10]);

      // force simulation code.
      var simulation = d3.forceSimulation()
        .force("link", d3.forceLink())
        .force("charge", d3.forceManyBody())
        .force("center", d3.forceCenter(width / 2, height / 2))
        .alphaTarget(.001)
        ;


      // SET UP LINKS

      var link = boxInner.append("g")
          .attr("class", "links") // add links class to lines we will create
        .selectAll("line")
        .data(d3Links)
        .enter().append("line")
          .attr("stroke-width", function(d) { return Math.sqrt(d.value); })
          .attr("stroke", function(d) { return color(d.relationship); })
          .attr("fill", "none")
          .attr("marker-end", "url(#end-triangle)") // attach arrow to lines.
        ;
      this.d3LinkVar = link;

      link.append("text")
          .text(function(d) {
            return d.relationship;
          })
          .attr("class", "d3-link-text")
          .attr('x', 6)
          .attr('y', 3);


      // build the arrow.
      // https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/marker-end
      // https://bl.ocks.org/d3noob/5141278
      box.append("svg:defs").selectAll("marker")
          .data(["end-triangle"])
        .enter().append("svg:marker") // marker element
          .attr("id", String)
          .attr("viewBox", "0 -5 10 10")
          .attr("refX", 35)
          .attr("refY", 0)
          .attr("markerWidth", 6)
          .attr("markerHeight", 6)
          .attr("markerUnits", "userSpaceOnUse") // https://stackoverflow.com/questions/48962654/how-to-maintain-the-svg-marker-width-and-height -- prevents arrow from resizing.
          .attr("orient", "auto")
        .append("svg:path") // define shape of triangle
          .attr("d", "M0,-5L10,0L0,5") // no fill.
          .attr("fill", "currentColor")
        ;

      


      // SET UP NODES

      var node = boxInner.append("g")
          .attr("class", "nodes") // add 'nodes' class to node groups we will create
        .selectAll("g")
        .data(d3Nodes)
        .enter().append("g").attr("class", "d3-node-group")
      this.d3NodeVar = node;

      // node.append("title")
        // .text(function(d) { return d.name; });

      const radius = 5;
      // var circles = // commented out so build stops crying this is an unused var.
      node.append("circle")
          .attr("r", 5)
          .attr("fill", function() { return nodeColor(7); })
          .call(d3.drag()
              .on("start", dragstarted)
              .on("drag", dragged)
              .on("end", dragended));

      // var labels = // commented out so build stops crying this is an unused var.
      node.append("text")
          .text(function(d) {
            return d.name;
          })
          .attr("class", "d3-node-text")
          .attr('x', 6)
          .attr('y', 3);


      // highlight related nodes & links
      node.on('click', function(event, sourceNode) {
        if ( sourceNode.index == this.clickedNodeIdx ) {
          this.clickedNodeIdx = -1;

          // reset attrs.
          node.selectAll("circle").attr("class", "");
          link.attr("class", "");
          return
        }
        else { this.clickedNodeIdx = sourceNode.index }

        //////// connected links & node indices
        var connectedNodeIndices = new Set(); // in an ideal world, we could just do a reduce, but we're already iterating
        connectedNodeIndices.add(sourceNode.index); // you are 'connected' to yourself.
        const connectedLinks = link.filter(function(singleLink) {
          // console.log(singleLink);
          const connected = sourceNode.index == singleLink.source.index || sourceNode.index == singleLink.target.index;
          if ( connected ) { 
            // only one of these will trigger, this is optimal over 'set add' checking the entire set.
            // note that if source index matches, then you add target ( & vice versa ) - add the side your sourceNode is not.
            if ( sourceNode.index == singleLink.source.index ) { connectedNodeIndices.add(singleLink.target.index ) }
            if ( sourceNode.index == singleLink.target.index ) { connectedNodeIndices.add(singleLink.source.index ) }
            return true;
          }
          return false;
        })
        // console.log(connectedLinks); // the set of attached links.

        ////// Assign Link Classes

        // Reset every class on links.
        link.attr("class", "");
        // add class to special links
        connectedLinks.attr("class", "d3-connected-link");

        ////// Assign Node Classes

        // get related nodes based on indices.
        const connectedNodes = node.filter(function(singleNode) {
          return connectedNodeIndices.has(singleNode.index);
        });

        // reset every class on circles
        node
          .selectAll("circle")
            .attr("class", "");
        // add class to connected circles
        connectedNodes
          .selectAll("circle")
            .attr("class", "d3-connected-node");



      }) // node on click.

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
      // ( 0, 0 is top left )
      function ticked() {
        node
            // control the circles...
            .attr("transform", function(d) {
              const dx = Math.max(radius + chartGutter/2, Math.min(width - radius - chartGutter/2, d.x));
              const dy = Math.max(radius + chartGutter/2, Math.min(height - radius - chartGutter/2, d.y));
              return "translate(" + dx + "," + dy + ")";
            })
            // control the surrounding groups...
            .attr("cx", function(d) { return d.x = Math.max(radius + chartGutter/2, Math.min(width - radius  - chartGutter/2, d.x)); })
            .attr("cy", function(d) { return d.y = Math.max(radius  + chartGutter/2, Math.min(height - radius - chartGutter/2, d.y)); });
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
