<script>
import {createApp} from 'vue';
import {SvgIcon} from '../svg.ts';
import {GET, DELETE} from '../modules/fetch.ts';

const sfc = {
  components: {SvgIcon},
  data() {
    return {
      loading: false,
      machineId: null,
      computedResult: '',
      canCompute: window.config.computeData.canCompute,
      machines: window.config.computeData.machines,
      textIsLoading: '<p>Process Computing...</p>',
      branch: window.config.computeData.branch,
      tag: window.config.computeData.tag,
      runs: window.config.computeData.runs,
      expandedRuns: [],
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
        ({value, done} = await reader.read());
        if (done) {
          return;
        }
        const chunkString = new TextDecoder().decode(value);
        this.computedResult += chunkString;
      }
    },
    toggleRun(index) {
      // Toggle the expanded state for the selected run
      if (this.expandedRuns.includes(index)) {
        this.expandedRuns = this.expandedRuns.filter((i) => i !== index);
      } else {
        this.expandedRuns.push(index);
      }
    },
    isExpanded(index) {
      // Check if a run is currently expanded
      return this.expandedRuns.includes(index);
    },

    compute() {
      // fetch html markup table
      const url = this.tag ? `${window.location.pathname}/execute?machineId=${this.machineId}&tag=${this.tag}` : `${window.location.pathname}/execute?machineId=${this.machineId}&branch=${this.branch}`;
      this.loading = true;
      this.computedResult = '';
      GET(url).then(async (res) => {
        if (!res.ok) {
          this.loading = false;
          return;
        }
        await this.readAllChunks(res.body);
        this.loading = false;
        this.computedResult = String(this.computedResult);
      });
    },

    deleteLog(filename) {
      const url = `${window.location.pathname}/delete-log?filename=${filename}`;
      DELETE(url).then(async (res) => {
        if (!res.ok) {
          this.loading = false;
          return;
        }
        window.location.reload();
      });
    },
  },
};

export function initRepoCompute() {
  const el = document.querySelector('#repo-compute-app');
  if (!el) return;
  createApp(sfc).mount(el);
}

export default sfc;
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

    <div class="gt-clear-both"/>

    <!-- Compute result will here -->
    <div v-if="computedResult !== '' || loading">
      <div
        class="markup content gt-overflow-x-scroll term-container gt-mt-3"
        v-html="textComputedResult"
      />
    </div>
  </div>

  <div class="empty center aligned" v-else>
    <SvgIcon name="octicon-server" :size="32"/>
    <h2>You can't compute for this repository.</h2>
  </div>

  <div class="run-list">
    <h2>Run Logs</h2>
    <div v-for="(run, index) in runs" :key="index" class="run-item">
      <div class="run-header" @click="toggleRun(index)">
        <span>📄 {{ run.filename }}</span>
        <span>{{ isExpanded(index) ? "▼" : "▶" }}</span>
        <span @click.prevent="deleteLog(run.filename)" style="cursor: pointer;">🗑️</span>
      </div>
      <div v-if="isExpanded(index)" class="run-content">
        <pre>{{ run.content }}</pre>
      </div>
    </div>
  </div>
</template>

<style>
@import "./../../css/terminal.css";

.run-list {
  margin: 20px auto;
  font-family: Arial, sans-serif;
}
.run-item {
  border: 1px solid #ddd;
  margin-bottom: 10px;
  border-radius: 4px;
}
.run-header {
  padding: 10px;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}
.run-content {
  padding: 10px;
  border-top: 1px solid #ddd;
  white-space: pre;
  overflow-x: auto;
}
</style>
