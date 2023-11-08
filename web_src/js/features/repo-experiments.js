import $ from 'jquery';

export function initExperimentsBranchChange() {
  $('#experiment-branch').on('change', (event) => {
    const url = `${window.location.pathname}?branch=${event.target.value}`;
    window.location.replace(url);
  });
}
