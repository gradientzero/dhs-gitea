<script>
import { createApp } from "vue";
import { SvgIcon } from "../svg.js";
import { GET } from "../modules/fetch.js";

const sfc = {
  components: { SvgIcon },
  data() {
    return {
      loading: false,
      machineId: null,
      computedResult: "",
      canCompute: window.config.computeData.canCompute,
      machines: window.config.computeData.machines,
      textIsLoading: "<p>Process Computing...</p>",
      branch: window.config.computeData.branch,
      tag: window.config.computeData.tag,
    };
  },

  computed: {
    textComputedResult() {
      if (!this.loading) return this.computedResult;
      return this.computedResult + this.textIsLoading;
    },
  },

  methods: {
    async readAllChunks(readableStream) {
      const reader = readableStream.getReader();
      let done, value;
      while (!done) {
        ({ value, done } = await reader.read());
        if (done) {
          return;
        }
        const chunkString = new TextDecoder().decode(value);
        this.computedResult += chunkString;
      }
    },

    compute: function () {
      // fetch html markup table
      const url = this.tag ? `${window.location.pathname}/execute?machineId=${this.machineId}&tag=${this.tag}` : `${window.location.pathname}/execute?machineId=${this.machineId}&branch=${this.branch}`;
      this.loading = true;
      this.computedResult = "";
      GET(url).then(async (res) => {
        if (!res.ok) {
          this.loading = false;
          return;
        }
        await this.readAllChunks(res.body);
        this.loading = false;
        this.computedResult += "";
      });
    },
  },
};

export default sfc;

export function initRepoCompute() {
  const el = document.getElementById("repo-compute-app");
  if (!el) return;
  const view = createApp(sfc);
  view.mount(el);
}
</script>

<template>
  <div class="ui attached segment gt-overflow-y-hidden" v-if="canCompute">
    <form
      class="ui form"
      method="post"
      @submit.prevent="compute"
      v-if="!loading"
    >
      <div class="field">
        <label>Machines</label>
        <select name="machine" class="ui fluid dropdown" v-model="machineId" required>
          <option :value="machine.ID" v-for="machine of machines">
            {{ machine.User }}@{{ machine.Host }}
          </option>
        </select>
      </div>
      <button class="ui primary button gt-float-right">Compute</button>
    </form>

    <div class="gt-clear-both"></div>

    <!-- Compute result will here -->
    <div v-if="computedResult !== '' || loading">
      <div
        class="markup content gt-overflow-x-scroll term-container gt-mt-3"
        v-html="textComputedResult"
      ></div>
    </div>
  </div>

  <div class="empty center aligned" v-else>
    <SvgIcon name="octicon-server" :size="32" />
    <h2>You can't compute for this repository.</h2>
  </div>
</template>

<style>
@import "./../../css/terminal.css";
</style>
