<script>
import { GET } from "../modules/fetch.js";

export default {
  data() {
    return {
      loading: true,
      selectedBranch: window.config.experimentData.selectedBranch,
      experimentHtmlTable: "",
    };
  },
  mounted() {
    // fetch html markup table
    const url = `${window.location.pathname}/table?branch=${this.selectedBranch}`;
    GET(url)
      .then((res) => res.json())
      .then((json) => {
        this.experimentHtmlTable = json.htmlTable;
        this.loading = false;
      });
  },
  unmounted() {},
};
</script>
<template>
  <div>
    <div class="ui segment gt-h-screen" v-if="loading">
      <div class="ui active dimmer">
        <div class="ui big text loader">Loading...</div>
      </div>
    </div>

    <div
      class="markup content gt-overflow-x-scroll"
      v-html="experimentHtmlTable"
      v-else
    ></div>
  </div>
</template>

<style scoped>
/* For some reason, .markup > table not working */
.markup > :first-child {
  display: table;
  width: 100%;
}
</style>
