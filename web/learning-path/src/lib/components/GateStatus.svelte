<script lang="ts">
  let { verified, manual, active }: { verified: boolean; manual: boolean; active: boolean } = $props();
</script>

{#if verified}
  <span class="status status-verified"><i></i> Test verified</span>
{:else if manual}
  <span class="status status-manual"><i></i> Manual note</span>
{:else if active}
  <span class="status status-active"><i></i> Current gate</span>
{:else}
  <span class="status status-pending"><i></i> Pending</span>
{/if}
