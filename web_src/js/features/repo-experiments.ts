import {createApp} from 'vue';
import RepoExperiments from '../components/RepoExperiments.vue';

export function initExperimentVue() {
  const experimentAppElement = document.querySelector('#experiment-app');
  const fileListView = createApp(RepoExperiments);
  fileListView.mount(experimentAppElement);
}
