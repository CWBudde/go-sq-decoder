# CBS/Logic-Steered Decoder Plan (TODO)

- [x] Define target CBS logic variant and references (CBS SQ, Tate/Fosgate concepts, AES papers).
- [x] Specify decode conventions (phase sign for Hilbert, channel naming, target separation metrics).
- [x] Implement CBS-style direction detectors using Lt/Rt: derive fronts/rears/diagonals (F=L+R, B=H(L)-H(R), D1=L-H(R), D2=H(L)-R) and measure short-term energy.
- [ ] Add band-splitting for steering control (e.g., 3 bands with independent envelopes; steer mids/highs more than lows).
- [x] Define logic control law: find dominant direction per band, apply boost to dominant channel(s) and attenuation to competitors; normalize to constant power and limit max gain change.
- [x] Add attack/release smoothing and hysteresis (e.g., fast attack ~5-10ms, release ~100-300ms) to avoid pumping and image jump.
- [x] Integrate steering into decoder pipeline (apply dynamic gains to matrix outputs).
- [x] Add tests for steering stability, separation improvement, and regression vs. basic matrix decode.
- [x] Add CLI flag/config for logic steering and document tradeoffs in README/docs.
