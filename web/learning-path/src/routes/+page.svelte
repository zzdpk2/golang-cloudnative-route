<script lang="ts">
  import { onMount } from "svelte";
  import { asset } from "$app/paths";
  import GateStatus from "$lib/components/GateStatus.svelte";
  import Icon from "$lib/components/Icon.svelte";
  import ProgressRing from "$lib/components/ProgressRing.svelte";
  import { allGates, gates, optionalGates, phases, type Gate, type PhaseId } from "$lib/journey";

  type ProgressResponse = {
    verified: string[];
    optionalVerified: string[];
    focus: Record<string, string>;
    updatedAt?: string;
    error?: string;
  };
  type ResetDispatchResponse = {
    queued?: boolean;
    phase?: number;
    actionsUrl?: string;
    syncCommand?: string;
    error?: string;
  };
  type PendingReset = {
    phase: PhaseId;
    publishedRevision: string;
  };
  type StatusFilter = "all" | "current" | "pending" | "noted" | "verified" | "optional";

  const manualStorageKey = "go-ddd-tdd.learning-path.v1";
  const legacyManualStorageKey = "go-ddd-tdd.learning-path.next.v1";
  const optionalManualStorageKey = "go-ddd-tdd.learning-path.optional.v1";
  const pendingResetStorageKey = "go-ddd-tdd.learning-path.pending-reset.v1";
  const statusFilters: StatusFilter[] = ["all", "current", "pending", "noted", "verified", "optional"];
  const pageSize = 6;
  const repositorySourceRoot =
    "https://github.com/zzdpk2/golang-cloudnative-route/blob/main";
  const resetApiBase = ((import.meta.env.VITE_RESET_API_URL as string | undefined) ?? "")
    .replace(/\/$/, "");
  const resetEndpoint = resetApiBase ? `${resetApiBase}/api/reset` : "";

  let verified = $state<Set<string>>(new Set());
  let manual = $state<Set<string>>(new Set());
  let optionalVerified = $state<Set<string>>(new Set());
  let optionalManual = $state<Set<string>>(new Set());
  let focus = $state<Record<string, string>>({});
  let evidenceAt = $state<string>();
  let lastReload = $state<Date>();
  let syncing = $state(true);
  let syncError = $state<string>();
  let phaseFilter = $state<PhaseId | "all">(1);
  let statusFilter = $state<StatusFilter>("all");
  let query = $state("");
  let page = $state(1);
  let selectedId = $state("F0");
  let copiedId = $state<string>();
  let resetDialog: HTMLDialogElement;
  let resetTarget = $state<PhaseId>();
  let resetAvailability = $state<"checking" | "available" | "unavailable">("checking");
  let resetting = $state(false);
  let resetError = $state<string>();
  let resetResult = $state<string>();
  let resetKey = $state("");
  let resetActionsUrl = $state<string>();
  let resetSyncCommand = $state("git pull --ff-only");
  let syncInFlight = false;
  let initialSelectionSet = false;

  let currentGate = $derived(gates.find((gate) => !verified.has(gate.id)));
  let manualOnlyCount = $derived(
    gates.filter((gate) => manual.has(gate.id) && !verified.has(gate.id)).length,
  );
  let overallPercent = $derived(Math.round((verified.size / gates.length) * 100));
  let optionalPercent = $derived(
    Math.round((optionalVerified.size / optionalGates.length) * 100),
  );
  let phaseProgress = $derived(
    phases.map((phase) => {
      const phaseGates = gates.filter((gate) => gate.phase === phase.id);
      const done = phaseGates.filter((gate) => verified.has(gate.id)).length;
      const phaseOptionalGates = optionalGates.filter((gate) => gate.phase === phase.id);
      const optionalDone = phaseOptionalGates.filter((gate) => optionalVerified.has(gate.id)).length;
      return {
        phase,
        gates: phaseGates,
        optionalGates: phaseOptionalGates,
        done,
        total: phaseGates.length,
        percent: Math.round((done / phaseGates.length) * 100),
        optionalDone,
      };
    }),
  );
  let currentPhaseProgress = $derived(
    currentGate ? phaseProgress.find((item) => item.phase.id === currentGate?.phase) : undefined,
  );
  let filteredGates = $derived.by(() => {
    const normalized = query.trim().toLowerCase();
    return allGates.filter((gate) => {
      if (phaseFilter !== "all" && gate.phase !== phaseFilter) return false;
      const isCurrent = currentGate?.id === gate.id;
      const isVerified = gateVerified(gate);
      const isNoted = gateManual(gate) && !isVerified;
      if (statusFilter === "current" && !isCurrent) return false;
      if (statusFilter === "pending" && (isVerified || isNoted || isCurrent)) return false;
      if (statusFilter === "noted" && !isNoted) return false;
      if (statusFilter === "verified" && !isVerified) return false;
      if (statusFilter === "optional" && !gate.optional) return false;
      if (!normalized) return true;
      return `${gate.id} ${gate.title} ${gate.summary} ${gate.tag}`
        .toLowerCase()
        .includes(normalized);
    });
  });
  let selected = $derived(
    allGates.find((gate) => gate.id === selectedId) ?? currentGate ?? gates[0]!,
  );
  let pageCount = $derived(Math.max(1, Math.ceil(filteredGates.length / pageSize)));
  let currentPage = $derived(Math.min(page, pageCount));
  let paginatedGates = $derived(
    filteredGates.slice((currentPage - 1) * pageSize, currentPage * pageSize),
  );
  let pageNumbers = $derived(Array.from({ length: pageCount }, (_, index) => index + 1));
  let visibleRange = $derived(
    filteredGates.length === 0
      ? "0"
      : `${(currentPage - 1) * pageSize + 1}–${Math.min(currentPage * pageSize, filteredGates.length)}`,
  );

  $effect(() => {
    const firstVisible = paginatedGates[0];
    if (firstVisible && !paginatedGates.some((gate) => gate.id === selectedId)) {
      selectedId = firstVisible.id;
    }
  });

  function gateVerified(gate: Gate) {
    return gate.optional ? optionalVerified.has(gate.id) : verified.has(gate.id);
  }

  function gateManual(gate: Gate) {
    return gate.optional ? optionalManual.has(gate.id) : manual.has(gate.id);
  }

  function sourceHref(source: string) {
    const encodedPath = source
      .split("/")
      .map((segment) => encodeURIComponent(segment))
      .join("/");

    return `${repositorySourceRoot}/${encodedPath}`;
  }

  async function syncProgress(showBusy = false) {
    if (syncInFlight) return;
    syncInFlight = true;
    if (showBusy) syncing = true;
    try {
      const response = await fetch(
        `${asset("/progress.json")}?t=${Date.now()}`,
        { cache: "no-store" },
      );
      const payload = (await response.json()) as ProgressResponse;
      if (!response.ok) throw new Error(payload.error || "progress unavailable");
      verified = new Set(payload.verified);
      optionalVerified = new Set(payload.optionalVerified || []);
      focus = payload.focus;
      evidenceAt = payload.updatedAt;
      lastReload = new Date();
      syncError = undefined;
      reconcilePendingReset(payload);
      if (!initialSelectionSet) {
        const nextGate = gates.find((gate) => !payload.verified.includes(gate.id));
        if (nextGate) {
          selectedId = nextGate.id;
          phaseFilter = nextGate.phase;
          page = 1;
        }
        initialSelectionSet = true;
      }
    } catch (error) {
      syncError = error instanceof Error ? error.message : "progress unavailable";
    } finally {
      syncInFlight = false;
      syncing = false;
    }
  }

  function toggleManual(id: string) {
    const gate = allGates.find((item) => item.id === id);
    if (!gate || gateVerified(gate)) return;
    const isOptional = Boolean(gate.optional);
    const next = new Set(isOptional ? optionalManual : manual);
    next.has(id) ? next.delete(id) : next.add(id);
    if (isOptional) {
      optionalManual = next;
    } else {
      manual = next;
    }
    try {
      localStorage.setItem(
        isOptional ? optionalManualStorageKey : manualStorageKey,
        JSON.stringify([...next]),
      );
    } catch {
      // The note remains available for this browser session.
    }
  }

  async function copyCommand(gate: Gate) {
    try {
      await navigator.clipboard.writeText(gate.command);
      copiedId = gate.id;
    } catch {
      copiedId = `error:${gate.id}`;
    }
    window.setTimeout(() => (copiedId = undefined), 1600);
  }

  function openDetail(id: string, forceReveal = false) {
    selectedId = id;
    window.requestAnimationFrame(() => {
      if (!forceReveal && !window.matchMedia("(max-width: 1120px)").matches) return;
      const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
      document.getElementById("gate-detail")?.scrollIntoView({ behavior: reducedMotion ? "auto" : "smooth", block: "start" });
      document.getElementById("selected-gate-heading")?.focus({ preventScroll: true });
    });
  }

  function scrollToSection(id: string) {
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    document.getElementById(id)?.scrollIntoView({
      behavior: reducedMotion ? "auto" : "smooth",
      block: "start",
    });
  }

  function selectPhase(phase: PhaseId | "all") {
    phaseFilter = phase;
    page = 1;
  }

  function selectStatus(status: StatusFilter) {
    statusFilter = status;
    page = 1;
  }

  function goToPage(nextPage: number) {
    page = Math.max(1, Math.min(nextPage, pageCount));
    window.requestAnimationFrame(() => {
      const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
      const heading = document.getElementById("gate-results-heading");
      heading?.scrollIntoView({
        behavior: reducedMotion ? "auto" : "smooth",
        block: "start",
      });
      heading?.focus({ preventScroll: true });
    });
  }

  async function probeResetAvailability() {
    if (!resetEndpoint) {
      resetAvailability = "unavailable";
      return;
    }
    try {
      const response = await fetch(resetEndpoint, {
        cache: "no-store",
        headers: { accept: "application/json" },
      });
      const contentType = response.headers.get("content-type") ?? "";
      if (!response.ok || !contentType.includes("application/json")) throw new Error();
      const payload = (await response.json()) as { available?: boolean };
      resetAvailability = payload.available ? "available" : "unavailable";
    } catch {
      resetAvailability = "unavailable";
    }
  }

  function openResetDialog(phase: PhaseId) {
    if (resetAvailability !== "available" || syncing || !evidenceAt) return;
    resetTarget = phase;
    resetError = undefined;
    resetResult = undefined;
    resetKey = "";
    resetActionsUrl = undefined;
    resetDialog.showModal();
  }

  function clearChapterNotes(phase: PhaseId) {
    const requiredIds = new Set(gates.filter((gate) => gate.phase === phase).map((gate) => gate.id));
    const optionalIds = new Set(optionalGates.filter((gate) => gate.phase === phase).map((gate) => gate.id));
    manual = new Set([...manual].filter((id) => !requiredIds.has(id)));
    optionalManual = new Set([...optionalManual].filter((id) => !optionalIds.has(id)));
    try {
      localStorage.setItem(manualStorageKey, JSON.stringify([...manual]));
      localStorage.setItem(optionalManualStorageKey, JSON.stringify([...optionalManual]));
    } catch {
      // The reset evidence remains authoritative even when storage is unavailable.
    }
  }

  function rememberPendingReset(phase: PhaseId) {
    try {
      const stored = JSON.parse(localStorage.getItem(pendingResetStorageKey) ?? "[]") as unknown;
      const pending = Array.isArray(stored)
        ? stored.filter((item): item is PendingReset =>
            typeof item === "object" && item !== null &&
            [1, 2, 3, 4].includes((item as PendingReset).phase) &&
            typeof (item as PendingReset).publishedRevision === "string")
        : [];
      const next = pending.filter((item) => item.phase !== phase);
      next.push({ phase, publishedRevision: evidenceAt! });
      localStorage.setItem(
        pendingResetStorageKey,
        JSON.stringify(next),
      );
    } catch {
      // The workflow still runs; only automatic note cleanup is unavailable.
    }
  }

  function reconcilePendingReset(payload: ProgressResponse) {
    if (!payload.updatedAt) return;
    try {
      const stored = JSON.parse(localStorage.getItem(pendingResetStorageKey) ?? "[]") as unknown;
      if (!Array.isArray(stored)) return;
      const remaining: PendingReset[] = [];
      for (const item of stored) {
        if (typeof item !== "object" || item === null) continue;
        const pending = item as PendingReset;
        if (![1, 2, 3, 4].includes(pending.phase) || typeof pending.publishedRevision !== "string") continue;
        if (payload.updatedAt === pending.publishedRevision) {
          remaining.push(pending);
          continue;
        }
        const requiredIds = gates.filter((gate) => gate.phase === pending.phase).map((gate) => gate.id);
        const optionalIds = optionalGates.filter((gate) => gate.phase === pending.phase).map((gate) => gate.id);
        const resetIsPublished = requiredIds.every((id) => !payload.verified.includes(id)) &&
          optionalIds.every((id) => !(payload.optionalVerified ?? []).includes(id));
        if (resetIsPublished) clearChapterNotes(pending.phase);
        else remaining.push(pending);
      }
      if (remaining.length > 0) localStorage.setItem(pendingResetStorageKey, JSON.stringify(remaining));
      else localStorage.removeItem(pendingResetStorageKey);
    } catch {
      // Ignore malformed or unavailable browser storage.
    }
  }

  async function resetChapter() {
    if (!resetTarget || resetting || resetKey.length < 32) return;
    resetting = true;
    resetError = undefined;
    const phase = resetTarget;
    try {
      const response = await fetch(resetEndpoint, {
        method: "POST",
        headers: {
          "content-type": "application/json",
          accept: "application/json",
          authorization: `Bearer ${resetKey}`,
        },
        body: JSON.stringify({ phase, confirmed: true }),
      });
      const contentType = response.headers.get("content-type") ?? "";
      const payload = contentType.includes("application/json")
        ? (await response.json()) as ResetDispatchResponse
        : {};
      if (!response.ok || !payload.queued) {
        throw new Error(payload.error || `reset service returned ${response.status}`);
      }
      rememberPendingReset(phase);
      resetActionsUrl = payload.actionsUrl;
      resetSyncCommand = payload.syncCommand ?? "git pull --ff-only";
      resetResult = "Reset queued. GitHub Actions will commit the starter, recalculate test evidence, and deploy the updated page.";
    } catch (error) {
      resetError = error instanceof Error ? error.message : "chapter reset could not be queued";
    } finally {
      resetKey = "";
      resetting = false;
    }
  }

  onMount(() => {
    try {
      const encoded =
        localStorage.getItem(manualStorageKey) ??
        localStorage.getItem(legacyManualStorageKey) ??
        "[]";
      const stored = JSON.parse(encoded) as unknown;
      const validIds = new Set(gates.map((gate) => gate.id));
      manual = new Set(
        Array.isArray(stored)
          ? stored.filter((id): id is string => typeof id === "string" && validIds.has(id))
          : [],
      );
      const optionalStored = JSON.parse(
        localStorage.getItem(optionalManualStorageKey) ?? "[]",
      ) as unknown;
      const optionalIds = new Set(optionalGates.map((gate) => gate.id));
      optionalManual = new Set(
        Array.isArray(optionalStored)
          ? optionalStored.filter(
              (id): id is string => typeof id === "string" && optionalIds.has(id),
            )
          : [],
      );
    } catch {
      manual = new Set();
      optionalManual = new Set();
    }
    void probeResetAvailability();
    void syncProgress(true);
    const interval = window.setInterval(() => {
      if (!document.hidden) void syncProgress(false);
    }, 5000);
    return () => window.clearInterval(interval);
  });
</script>

<svelte:head>
  <title>Commerce → llm-d Router · Learning Path</title>
  <meta name="description" content="An interactive, test-synchronized progress map for the Commerce to llm-d Router journey." />
</svelte:head>



<main class="learning-world">
  <div class="ambient ambient-pink" aria-hidden="true"></div>
  <div class="ambient ambient-green" aria-hidden="true"></div>

  <header class="top-dock">
    <button class="wordmark" onclick={() => scrollToSection("overview")} aria-label="Return to overview">
      <span class="wordmark-mark">R</span>
      <span><strong>Router path</strong><small>Commerce → llm-d</small></span>
    </button>
    <nav aria-label="Page sections">
      <button onclick={() => scrollToSection("overview")}>Overview</button>
      <button onclick={() => scrollToSection("journey")}>Journey</button>
      <button onclick={() => scrollToSection("labs")}>Lab deck</button>
    </nav>
    <div class="dock-progress" role="progressbar" aria-valuemin="0" aria-valuemax={gates.length} aria-valuenow={verified.size} aria-label={`${overallPercent}% of the journey is test verified`}>
      <span><i style:width={`${overallPercent}%`}></i></span><strong>{overallPercent}%</strong>
    </div>
    <div class:error={Boolean(syncError)} class="sync-state" aria-live="polite" aria-atomic="true">
      <i></i><span>{syncError ? "Evidence offline" : lastReload ? "Tests connected" : "Connecting…"}</span>
    </div>
    <button class="sync-button" aria-label="Sync test progress now" onclick={() => void syncProgress(true)} disabled={syncing}>
      <span class:spin={syncing}><Icon name="refresh" /></span><span>{syncing ? "Syncing" : "Sync progress"}</span>
    </button>
  </header>

  <section class="hero-deck" id="overview">
    <div class="hero-copy">
      <span class="eyebrow"><i></i> Production learning route</span>
      <h1>Learn the system.<br /><em>Earn the signal.</em></h1>
      <p>One deliberate path from Go fundamentals and commerce architecture to Kubernetes, observability, and the llm-d Router.</p>
      <div class="hero-chips">
        <span><strong>{gates.length}</strong> focused gates</span>
        <span><strong>{verified.size}</strong> test verified</span>
        <span><strong>{manualOnlyCount}</strong> manual notes</span>
        <span class="optional-chip"><strong>★ {optionalVerified.size}/{optionalGates.length}</strong> optional · excluded from total</span>
      </div>
      <button class="hero-cta" onclick={() => scrollToSection("journey")}>Explore the route <Icon name="arrow" /></button>
    </div>

    <div class="progress-orbit">
      <div class="orbit-ring orbit-outer" aria-hidden="true"><i></i><b></b></div>
      <div class="orbit-ring orbit-inner" aria-hidden="true"><i></i></div>
      <ProgressRing done={verified.size} total={gates.length} />
      <span class="orbit-caption">test-verified</span>
    </div>

    {#if currentGate}
      <article class="mission-ticket">
        <div class="ticket-edge" aria-hidden="true"></div>
        <div class="ticket-top"><span><i></i> Current mission</span><strong>{currentGate.id}</strong></div>
        <span class="ticket-phase">Phase 0{currentGate.phase} · {phases[currentGate.phase - 1]!.shortTitle}</span>
        <h2>{currentGate.title}</h2>
        <p>{currentGate.summary}</p>
        {#if currentPhaseProgress}
          <div class="ticket-meter" role="progressbar" aria-valuemin="0" aria-valuemax={currentPhaseProgress.total} aria-valuenow={currentPhaseProgress.done} aria-label={`${currentPhaseProgress.percent}% of the current phase is test verified`}><span><i style:width={`${currentPhaseProgress.percent}%`}></i></span><small>{currentPhaseProgress.done}/{currentPhaseProgress.total} phase gates verified</small></div>
        {/if}
        <button onclick={() => openDetail(currentGate.id, true)}>Open mission <Icon name="arrow" /></button>
      </article>
    {:else}
      <article class="mission-ticket complete-ticket">
        <div class="ticket-top"><span><i></i> Journey complete</span><strong><Icon name="check" /></strong></div>
        <h2>Upstream capstone unlocked</h2>
        <p>Continue with the pinned llm-d Router reading map and one upstream issue.</p>
      </article>
    {/if}
  </section>

  <section class="journey-stage" id="journey" aria-labelledby="journey-heading">
    <div class="section-heading">
      <div><span class="eyebrow"><i></i> One route, four worlds</span><h2 id="journey-heading">The learning constellation</h2></div>
      <div class="legend" aria-label="Gate status legend"><span class="legend-verified"><i></i>Verified</span><span class="legend-current"><i></i>Current</span><span class="legend-noted"><i></i>Noted</span><span class="legend-optional"><i></i>Optional · skippable</span></div>
    </div>
    <div class="journey-tree">
      <div class="tree-spine" aria-hidden="true"><i style:height={`${overallPercent}%`}></i><b></b></div>
      {#each phaseProgress as item, index (item.phase.id)}
        <div class:branch-right={index % 2 === 1} class="tree-branch">
          <span class:complete={item.percent === 100} class:active={currentGate?.phase === item.phase.id} class="tree-node" aria-hidden="true">0{item.phase.id}</span>
          <button class:active={phaseFilter === item.phase.id} aria-label={`Show ${item.phase.title}: ${item.percent}% test verified`} aria-pressed={phaseFilter === item.phase.id} onclick={() => { selectPhase(item.phase.id); scrollToSection("labs"); }}>
            <span class="phase-card-top"><span>{item.phase.eyebrow}</span><strong>{item.percent}%</strong></span>
            <h3>{item.phase.title}</h3><p>{item.phase.description}</p>
            <div class="tree-gates" aria-label={`${item.total} gates in ${item.phase.title}`}>
              {#each item.gates as gate (gate.id)}
                <span class:verified={verified.has(gate.id)} class:current={currentGate?.id === gate.id} class:noted={manual.has(gate.id) && !verified.has(gate.id)} title={`${gate.id} · ${gate.title}`}>{gate.id}</span>
              {/each}
            </div>
            {#if item.optionalGates.length > 0}
              <div class="tree-optionals" aria-label={`${item.optionalGates.length} skippable optional groups in ${item.phase.title}`}>
                <span class="optional-route-label">★ Optional branch · {item.optionalDone}/{item.optionalGates.length} verified · does not block</span>
                <div>
                  {#each item.optionalGates as gate (gate.id)}
                    <span class:verified={optionalVerified.has(gate.id)} class:noted={optionalManual.has(gate.id) && !optionalVerified.has(gate.id)} title={`${gate.id} · ${gate.title}`}>{gate.id.replace("OPT-", "")}</span>
                  {/each}
                </div>
              </div>
            {/if}
            <span class="tree-card-footer"><span>{item.done} of {item.total} verified</span><span class="phase-card-meter"><i style:width={`${item.percent}%`}></i></span></span>
          </button>
        </div>
      {/each}
    </div>
  </section>

  <section class="lab-stage" id="labs" aria-labelledby="labs-heading">
    <div class="section-heading lab-heading">
      <div><span class="eyebrow"><i></i> Practice workspace</span><h2 id="labs-heading">The gate deck</h2></div>
      <div class="deck-counts">
        <span class="result-count">{filteredGates.length} shown · {gates.length} required</span>
        <div class="optional-progress" role="progressbar" aria-valuemin="0" aria-valuemax={optionalGates.length} aria-valuenow={optionalVerified.size} aria-label={`${optionalPercent}% of optional groups are test verified`}>
          <span><i style:width={`${optionalPercent}%`}></i></span><small>★ {optionalVerified.size}/{optionalGates.length} optional · +0% required</small>
        </div>
      </div>
    </div>
    <div class="chapter-shelf" aria-label="Choose a learning chapter">
      {#each phaseProgress as item (item.phase.id)}
        <div class:active={phaseFilter === item.phase.id} class="chapter-card-shell">
          <button class="chapter-select" aria-pressed={phaseFilter === item.phase.id} aria-label={`${item.phase.title}: ${item.percent}% required, ${item.optionalDone} of ${item.optionalGates.length} optional verified`} onclick={() => selectPhase(item.phase.id)}>
            <span class="chapter-index">0{item.phase.id}</span>
            <span class="chapter-copy"><strong>{item.phase.shortTitle}</strong><small>{item.total} required{item.optionalGates.length ? ` · ${item.optionalGates.length} optional` : ""}</small></span>
            <span class="chapter-score"><strong>{item.percent}%</strong><small>{item.done}/{item.total}</small></span>
            <span class="chapter-meter"><i style:width={`${item.percent}%`}></i></span>
          </button>
          <button class="chapter-reset" disabled={resetAvailability !== "available" || syncing || !evidenceAt} title={resetAvailability !== "available" ? "Remote reset is not configured" : syncing || !evidenceAt ? "Wait for the first progress sync" : `Reset ${item.phase.title} through GitHub Actions`} aria-label={`Reset ${item.phase.title} code and progress`} onclick={() => openResetDialog(item.phase.id)}>
            {resetAvailability === "checking" ? "Checking…" : resetAvailability !== "available" ? "Reset unavailable" : syncing || !evidenceAt ? "Waiting for evidence…" : "Reset chapter"}
          </button>
        </div>
      {/each}
      <button class:active={phaseFilter === "all"} class="all-chapters" aria-pressed={phaseFilter === "all"} onclick={() => selectPhase("all")}><span>∞</span><strong>All chapters</strong><small>Cross-chapter search</small></button>
    </div>
    <div class="filter-bar">
      <label class="search-box"><Icon name="search" /><input aria-label="Search learning gates" bind:value={query} oninput={() => (page = 1)} placeholder="Search a topic, tag, or gate…" /></label>
      <div class="filter-tabs" aria-label="Filter gates by status">
        {#each statusFilters as status}
          <button class:active={statusFilter === status} aria-pressed={statusFilter === status} onclick={() => selectStatus(status)}>{status}</button>
        {/each}
      </div>
    </div>
    <div class="lab-layout">
      <div class="gate-results">
        <div class="results-heading" id="gate-results-heading" tabindex="-1">
          <div><span>{phaseFilter === "all" ? "All chapters" : phases[phaseFilter - 1]!.eyebrow}</span><h3>{phaseFilter === "all" ? "Cross-chapter results" : phases[phaseFilter - 1]!.title}</h3></div>
          <p aria-live="polite" aria-atomic="true"><strong>{visibleRange}</strong> of {filteredGates.length} · page {currentPage}/{pageCount}</p>
        </div>
        <div class="gate-grid">
        {#each paginatedGates as gate, index (gate.id)}
          {@const isVerified = gateVerified(gate)}
          {@const isManual = gateManual(gate) && !isVerified}
          {@const isCurrent = currentGate?.id === gate.id}
          <article class:optional={gate.optional} class:verified={isVerified} class:selected={selected.id === gate.id} class:current={isCurrent} class="gate-card" style={`--delay: ${Math.min(index, 12) * 35}ms`}>
            <div class="gate-card-top"><span class="gate-code">{gate.id}</span><span class="tag">{gate.tag}</span>{#if gate.optional}<span class="optional-badge">★ Optional</span>{/if}<GateStatus verified={isVerified} manual={isManual} active={isCurrent} /></div>
            <h3>{gate.title}</h3><p>{gate.summary}</p>
            <div class="gate-meter" role="progressbar" aria-valuemin="0" aria-valuemax="100" aria-valuenow={isVerified ? 100 : 0} aria-label={isVerified ? "100% test verified" : "0% test verified"}><span><i style:width={isVerified ? "100%" : "0%"}></i></span><small>{isVerified ? "contract passed" : "contract open"}</small></div>
            {#if gate.optional}<p class="skip-note">Skippable · never changes required progress</p>{/if}
            <div class="gate-card-actions">
              <button class="note-button" aria-label={isVerified ? `${gate.id} test verified` : `Toggle manual note for ${gate.id}`} aria-pressed={isManual} onclick={() => toggleManual(gate.id)} disabled={isVerified}><Icon name="check" />{isVerified ? "Verified" : isManual ? "Noted" : "Add note"}</button>
              <button class="open-button" aria-label={`Open ${gate.id} ${gate.title}`} onclick={() => openDetail(gate.id)}>Open <Icon name="arrow" /></button>
            </div>
          </article>
        {/each}
        {#if filteredGates.length === 0}
          <div class="empty-state"><span>∅</span><h3>No gates found</h3><p>Try another phrase or reset the filters.</p><button onclick={() => { query = ""; selectPhase(currentGate?.phase ?? 1); selectStatus("all"); }}>Reset view</button></div>
        {/if}
        </div>
        {#if filteredGates.length > 0}
          <nav class="pagination" aria-label="Gate pages">
            <button class="page-arrow" onclick={() => goToPage(currentPage - 1)} disabled={currentPage === 1} aria-label="Previous gate page">← <span>Previous</span></button>
            <div>
              {#each pageNumbers as pageNumber}
                <button class:active={currentPage === pageNumber} aria-current={currentPage === pageNumber ? "page" : undefined} aria-label={`Open gate page ${pageNumber}`} onclick={() => goToPage(pageNumber)}>{pageNumber}</button>
              {/each}
            </div>
            <button class="page-arrow" onclick={() => goToPage(currentPage + 1)} disabled={currentPage === pageCount} aria-label="Next gate page"><span>Next</span> →</button>
          </nav>
        {/if}
      </div>
      {#if paginatedGates.length > 0}
      <aside class="detail-panel" id="gate-detail">
        <div class="detail-sticky">
          <div class="detail-top"><span class:optional={selected.optional} class="detail-id">{selected.optional ? "★" : selected.id}</span><GateStatus verified={gateVerified(selected)} manual={gateManual(selected) && !gateVerified(selected)} active={currentGate?.id === selected.id} /></div>
          <span class="eyebrow"><i></i> {selected.optional ? "Skippable extension" : "Selected mission"}</span>
          <h2 id="selected-gate-heading" tabindex="-1">{selected.title}</h2><p class="detail-summary">{selected.summary}</p>
          <div class="contract-progress">
            <div><span>{selected.optional ? "Optional test evidence" : "Test evidence"}</span><strong>{gateVerified(selected) ? "100%" : "0%"}</strong></div>
            <span><i style:width={gateVerified(selected) ? "100%" : "0%"}></i></span><p>{selected.optional ? "Tracked independently. You may skip this branch without blocking the next required gate." : "Binary by design: this gate closes only when its focused contract passes."}</p>
          </div>
          <div class="detail-block focus-block"><span>{selected.optional ? "Optional review lens" : "100 Go Mistakes lens"}</span><p>{selected.optional ? selected.focus : focus[selected.id] || "Run the exercise helper to publish this gate's review focus."}</p></div>
          <div class="detail-block command-block">
            <div><span>Focused command</span><button onclick={() => void copyCommand(selected)}><Icon name={copiedId === selected.id ? "check" : "copy"} />{copiedId === selected.id ? "Copied" : copiedId === `error:${selected.id}` ? "Copy failed" : "Copy"}</button></div>
            <code>{selected.command}</code>
          </div>
          <button class="detail-note" aria-pressed={gateManual(selected) && !gateVerified(selected)} onclick={() => toggleManual(selected.id)} disabled={gateVerified(selected)}><Icon name="check" /> {gateVerified(selected) ? "Verified by tests" : gateManual(selected) ? "Remove manual note" : "Mark manual note"}</button>
          <div class:optional={selected.optional} class="evidence-note"><i></i><p><strong>{selected.optional ? "A branch, not a barrier." : "Evidence, not vibes."}</strong> Manual notes stay in this browser. Run <code>{selected.optional ? "go run ./cmd/exercise optional" : "go run ./cmd/exercise next"}</code> and sync to publish test evidence.</p></div>
          <footer class="detail-footer"><span>Last evidence sync</span><time>{evidenceAt ? new Date(evidenceAt).toLocaleString() : lastReload ? lastReload.toLocaleString() : "Waiting for runner"}</time></footer>
        </div>
      </aside>
      {/if}
    </div>
  </section>

  <footer class="page-footer">
    <div><span class="wordmark-mark">R</span><p><strong>Router learning path</strong><small>A test-led route into production inference infrastructure.</small></p></div>
    <nav aria-label="Learning references"><a href={sourceHref("docs/LEARNING_PATH.md")} target="_blank" rel="noreferrer"><Icon name="source" />Learning guide</a><a href={sourceHref("docs/go100/README.md")} target="_blank" rel="noreferrer"><span class="book-icon">100</span>Go Mistakes clinics</a></nav>
  </footer>

  <dialog class="reset-dialog" bind:this={resetDialog} onclose={() => { if (!resetting) { resetTarget = undefined; resetKey = ""; } }}>
    {#if resetTarget}
      {@const target = phases[resetTarget - 1]!}
      <div class="reset-dialog-mark" aria-hidden="true">0{resetTarget}</div>
      <span class="eyebrow"><i></i> Protected repository action</span>
      <h2>Restart {target.shortTitle}?</h2>
      <p>GitHub Actions will replace this chapter with its red starter and commit the result to <code>main</code>. The next Pages deployment will return this chapter's required and optional progress to zero.</p>
      <p class="reset-dialog-safety"><strong>Your work stays in Git history.</strong> Contract tests and other chapters are not changed.</p>
      {#if !resetResult}
        <label class="reset-key-field">
          <span>Reset Key</span>
          <input type="password" bind:value={resetKey} autocomplete="off" spellcheck="false" placeholder="Paste the 32+ character Vercel secret" disabled={resetting} />
          <small>The key is sent only to the configured Vercel API and is never stored by this page.</small>
        </label>
      {/if}
      {#if resetError}<p class="reset-message error" role="alert">{resetError}</p>{/if}
      {#if resetResult}
        <p class="reset-message success" role="status">{resetResult}</p>
        <div class="reset-next-step"><span>After the workflow completes</span><code>{resetSyncCommand}</code>{#if resetActionsUrl}<a href={resetActionsUrl} target="_blank" rel="noreferrer">Open GitHub Actions ↗</a>{/if}</div>
      {/if}
      <div class="reset-dialog-actions">
        <button onclick={() => resetDialog.close()} disabled={resetting}>{resetResult ? "Close" : "Keep my work"}</button>
        {#if !resetResult}<button class="confirm-reset" onclick={() => void resetChapter()} disabled={resetting || resetKey.length < 32}>{resetting ? "Queueing reset…" : "Queue reset"}</button>{/if}
      </div>
    {/if}
  </dialog>
</main>
