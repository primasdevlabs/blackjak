export interface SequencedEvent {
  sequence?: number;
  [key: string]: any;
}

export class EventReorderBuffer<T extends SequencedEvent> {
  private expectedSequence = 1;
  private buffer: Map<number, T> = new Map();
  private maxBufferSize = 500;

  constructor(initialSequence = 1) {
    this.expectedSequence = initialSequence;
  }

  /**
   * Pushes an incoming event into the reorder buffer.
   * Returns ordered array of events ready for processing.
   */
  public push(event: T): T[] {
    const seq = event.sequence;
    if (seq === undefined || seq <= 0) {
      // Unsequenced event: return immediately
      return [event];
    }

    if (seq < this.expectedSequence) {
      // Duplicate or old event, discard
      console.warn(`[EventReorderBuffer] Discarding duplicate/old event sequence #${seq}, expected >= #${this.expectedSequence}`);
      return [];
    }

    this.buffer.set(seq, event);

    // Evict oldest if buffer overflows
    if (this.buffer.size > this.maxBufferSize) {
      const minSeq = Math.min(...Array.from(this.buffer.keys()));
      console.warn(`[EventReorderBuffer] Buffer capacity exceeded. Forcing flush from sequence #${minSeq}`);
      this.expectedSequence = minSeq;
    }

    const ready: T[] = [];
    while (this.buffer.has(this.expectedSequence)) {
      const nextEvent = this.buffer.get(this.expectedSequence)!;
      this.buffer.delete(this.expectedSequence);
      ready.push(nextEvent);
      this.expectedSequence++;
    }

    return ready;
  }

  public reset(nextExpectedSequence = 1): void {
    this.buffer.clear();
    this.expectedSequence = nextExpectedSequence;
  }

  public getPendingCount(): number {
    return this.buffer.size;
  }
}
