// i'm disabling this because there are a bunch of classes here that
// are pretty trivially different and having multiple files seemed silly
/* eslint-disable max-classes-per-file */

import s5event from './s5event';
import eventName from './eventname';
import eventBase from './eventbase';

export interface eventDescriptor {
  Generate(e: Event | null, x?: number, y?: number): s5event;
}

export default abstract class eventDescriptorBase implements eventDescriptor {
  // raw can be null, especially for user-generated events
  Generate(raw: Event | null, x?: number, y?: number): s5event {
    return this.createEvent(raw, x, y);
  }

  abstract createEvent(raw: Event | null, x?: number, y?: number): s5event;
}

export class buttonDown extends eventDescriptorBase {
  // eslint-disable-next-line class-methods-use-this
  createEvent(raw: Event | null, x?: number, y?: number): s5event {
    return new eventBase(eventName.MouseButtonDown, raw, x, y);
  }
}

export class buttonUp extends eventDescriptorBase {
  // eslint-disable-next-line class-methods-use-this
  createEvent(raw: Event | null, x?: number, y?: number): s5event {
    return new eventBase(eventName.MouseButtonUp, raw, x, y);
  }
}

export class mouseMove extends eventDescriptorBase {
  // eslint-disable-next-line class-methods-use-this
  createEvent(raw: Event | null, x?: number, y?: number): s5event {
    return new eventBase(eventName.MouseMove, raw, x, y);
  }
}

export class mouseEnter extends eventDescriptorBase {
  // eslint-disable-next-line class-methods-use-this
  createEvent(raw: Event | null, x?: number, y?: number): s5event {
    return new eventBase(eventName.MouseEnter, raw, x, y);
  }
}

export class mouseLeave extends eventDescriptorBase {
  // eslint-disable-next-line class-methods-use-this
  createEvent(raw: Event | null, x?: number, y?: number): s5event {
    return new eventBase(eventName.MouseLeave, raw, x, y);
  }
}
