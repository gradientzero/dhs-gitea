<script>
import {GET} from '../modules/fetch.ts';

export default {
  data() {
    return {
      loading: true,
      selectedBranch: window.config.experimentData.selectedBranch,
      selectedTag: window.config.experimentData.selectedTag,
      experimentHtmlTable: '',
    };
  },
  mounted() {
    // fetch html markup table
    const url = this.selectedTag ? `${window.location.pathname}/table?tag=${this.selectedTag}` : `${window.location.pathname}/table?branch=${this.selectedBranch}`;
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
    />
  </div>
</template>

<style scoped>
/* For some reason, .markup > table not working */
.markup > :first-child {
  display: table;
  width: 100%;
}
</style>
