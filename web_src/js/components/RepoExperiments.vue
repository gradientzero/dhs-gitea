<script>

import {GET} from '../modules/fetch.js';

export default {
  data() {
    return {
      loading: true,
      selectedBranch: window.config.experimentData.branch,
      experimentHtmlTable: "",
    };
  },
  computed: {
    title() {
      return window.config.experimentData.title;
    },
    branches() {
      return window.config.experimentData.branches;
    },
  },
  mounted() {
    // fetch html markup table
    const url = `${window.location.pathname}/table?branch=${this.selectedBranch}`
    GET(url)
      .then(res => res.json())
        .then((json) => {
      console.log(json)
      this.experimentHtmlTable = json.htmlTable;
      this.loading = false;
    });
  },
  unmounted() {
  },
  methods: {
    branchChange(event) {
      const url = `${window.location.pathname}?branch=${event.target.value}`;
      window.location.replace(url);
    },
  },
};
</script>
<template>
  <h2 class="ui dividing header gt-df gt-sb">
    <div>
      {{ title }}
    </div>

    <div class="gt-p-3">
      <select class="ui fluid dropdown" @change="branchChange($event)" v-model="selectedBranch">
        <option :value="branch" v-for="branch in branches">{{ branch }}</option>
      </select>
    </div>
  </h2>

  <!-- Main Content Here -->
  <div class="ui segment gt-h-screen" v-if="loading">
    <div class="ui active dimmer">
      <div class="ui big text loader">Loading...</div>
    </div>
  </div>

  <div class="markup content gt-overflow-x-scroll" v-html="experimentHtmlTable" v-else></div>

</template>

<style scoped>
  /* For some reason, .markup > table not working */
  .markup > :first-child {
    display: table;
    width: 100%;
  }
</style>
