<script>

import {createApp} from 'vue';
import {SvgIcon} from '../svg.js';
import {GET} from '../modules/fetch.js';


const sfc = {
  components: {SvgIcon},
  data(){
    return {
      loading: false,
      isComputed: false,
      machineId: null,
      computedResult: '',
      canCompute: window.config.computeData.canCompute,
      machines: window.config.computeData.machines,
    };
  },
  computed: {

  },
  methods: {
    compute: function () {
      // fetch html markup table
      const url = `${window.location.pathname}/execute?machineId=${this.machineId}`
      this.loading = true;
      GET(url)
        .then(res => res.json())
        .then((json) => {
          this.computedResult = json.result;

          this.loading = false;
          this.isComputed = true;
        });
    }
  }
};

export default sfc;

export function initRepoCompute(){
  const el = document.getElementById("repo-compute-app");
  if(!el) return;
  const view = createApp(sfc);
  view.mount(el);
}
</script>

<template>

  <div class="ui attached segment gt-overflow-y-hidden" v-if="canCompute">
    <form class="ui form" method="post" @submit.prevent="compute">
      <div class="field">
        <label>Machines</label>
        <select name="machine" class="ui fluid dropdown" v-model="machineId">
          <option :value="machine.ID" v-for="machine of machines">{{ machine.User }}@{{ machine.Host }}</option>
        </select>
      </div>

      <button class="ui primary button gt-float-right">
        Compute
      </button>
    </form>

    <div class="gt-clear-both"></div>

    <!-- Compute result will here -->
    <div class="gt-h-screen" v-if="loading">
      <div class="ui active dimmer">
        <div class="ui big text loader">Computing...</div>
      </div>
    </div>
    <div v-else-if="isComputed">
      <div class="markup content gt-overflow-x-scroll">{{ this.computedResult }}</div>
    </div>
  </div>

  <div class="empty center aligned" v-else>
    <SvgIcon name="octicon-server" :size=32 />
    <h2>You can't compute for this repository.</h2>
  </div>
</template>
