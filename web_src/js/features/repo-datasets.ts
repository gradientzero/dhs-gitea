import $ from 'jquery';

export function initDatasetsBranchChange() {
  $('#dataset-branch').on('change', (event: any) => {
    const url = `${window.location.pathname}?branch=${event.target.value}`;
    window.location.replace(url);
  });
}
