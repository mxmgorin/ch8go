class SimpleProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
    this.buf = null;
    this.port.onmessage = (e) => (this.buf = e.data);
  }

  process(_, outputs) {
    const out = outputs[0][0];

    if (this.buf)
      out.set(this.buf);
    else
      out.fill(0);

    this.port.postMessage("need");
    return true;
  }
}

registerProcessor("simple-processor", SimpleProcessor);
