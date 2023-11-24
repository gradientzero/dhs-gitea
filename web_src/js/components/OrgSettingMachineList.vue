<script>

import {createApp} from 'vue';
import {SvgIcon} from '../svg.js';

const sfc = {
  components: {SvgIcon},
  data(){
    return {
      machines: window.config.settingsMachineData.machines,
      csrfToken: window.config.settingsMachineData.csrfToken,
      link: window.config.settingsMachineData.link,
      selectMachine: {}, // empty by default
    };
  },
  computed:{
    deleteLink() {
      return this.$data.link + '/delete';
    }
  },
  methods: {

    editLink(machine) {
      return this.$data.link + '/edit?id=' + machine.ID
    },
    deleteMachine(event) {
      event.preventDefault();

      $('#machine-delete-modal')
        .modal({
          onApprove: function() {
            event.target.submit();
          }
        })
        .modal('show');
    }
  }
};

export default sfc;

export function initOrgSettingMachineList(){
  const el = document.getElementById('setting-machine-app');
  if (!el) return;

  const view = createApp(sfc);
  view.mount(el);
}
</script>
<template>

    <div v-if="!machines.length">
      No machine has been setup yet.
    </div>

    <div v-else class="flex-list">
      <div class="flex-item" v-for="machine in machines">

        <div class="flex-item-leading text green">
          <SvgIcon name="octicon-server" :size=32 />
        </div>

        <div class="flex-item-main">
          <div class="flex-item-title">{{machine.Name}}</div>
          <div class="flex-item-body">
            {{machine.User}}@{{ machine.Host }}
          </div>
        </div>

        <div class="flex-item-trailing">
          <a :href="editLink(machine)" class="ui tiny button">
            <SvgIcon name="octicon-file" />
          </a>
          <form class="ui form" :action="deleteLink" method="post" @submit="deleteMachine($event)">
            <input type="hidden" name="_csrf" :value="csrfToken">
            <input name="id" type="hidden" :value="machine.ID">
            <button type="submit" class="ui red tiny button">
              <SvgIcon name="octicon-trash" />
            </button>
          </form>
        </div>

      </div>

<!--      Modal to delete machine-->
      <div class="ui modal" id="machine-delete-modal">
        <div class="header">Ssh Key Delete</div>
        <div class="content">
          <p>Are you sure to delete machine?</p>
        </div>
        <div class="actions">
          <div class="ui positive button">Delete</div>
          <div class="ui negative button">Cancel</div>
        </div>
      </div>

    </div>
</template>
