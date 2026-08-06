<script lang="ts">
  let { done, total }: { done: number; total: number } = $props();
  const radius = 50;
  const circumference = 2 * Math.PI * radius;
  let percent = $derived(total === 0 ? 0 : Math.round((done / total) * 100));
  let offset = $derived(circumference - (percent / 100) * circumference);
</script>

<div class="progress-ring" role="progressbar" aria-valuemin="0" aria-valuemax={total} aria-valuenow={done} aria-label={`${done} of ${total} gates test verified, ${percent}%`}>
  <svg viewBox="0 0 120 120">
    <circle class="ring-track" cx="60" cy="60" r={radius}></circle>
    <circle class="ring-value" cx="60" cy="60" r={radius} stroke-dasharray={circumference} stroke-dashoffset={offset}></circle>
  </svg>
  <div><strong>{percent}%</strong><span>{done}/{total}</span></div>
</div>
