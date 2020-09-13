/* eslint-disable no-console */
import {Event as BusAnyEvent, Event as BusEvent} from "../../genproto/farm_ng_proto/tractor/v1/io";
import { BusEventEmitter } from "./BusEventEmitter";
import { MediaStreamEmitter } from "./MediaStreamEmitter";

export function getWebRTCEmitters(
  endpoint: string
): [BusEventEmitter, MediaStreamEmitter] {
  const busEventEmitter = new BusEventEmitter();
  const mediaStreamEmitter = new MediaStreamEmitter();
  const pc = new RTCPeerConnection({
    iceServers: [] // no STUN servers, since we only support LAN
  });

  // TODO: jin remove
  busEventEmitter.on("ipc/announcement/webrtc-proxy", (event: BusAnyEvent) => {
    console.log("invoking bus event callback")
    console.log(event);
  });

  pc.ontrack = (event) => {
    if (event.track.kind != "video") {
      console.log(
        `${event.track.kind} track added, but only video tracks are supported.`
      );
      return;
    }
    mediaStreamEmitter.addVideoStream(event.streams[0]);
  };

  // pc.oniceconnectionstatechange = (_: Event) =>
  //   console.log(`New ICE Connection state: ${pc.iceConnectionState}`);

  // TODO: debug
  pc.oniceconnectionstatechange = (_: Event) => {
    console.log("ice connection state change event", _);
    console.log(`New ICE Connection state: ${pc.iceConnectionState}`);
  };

  pc.onicecandidate = async (event: RTCPeerConnectionIceEvent) => {
    if (event.candidate === null) {
      const response = await fetch(endpoint, {
        method: "POST",
        mode: "cors",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          sdp: btoa(JSON.stringify(pc.localDescription))
        })
      });

      // TODO: @jin
      var abc = JSON.stringify({
        sdp: btoa(JSON.stringify(pc.localDescription))
      }, null, 2);
      console.log(abc);
      console.log("getting ice candidate");

      const responseJson = await response.json();
      try {
        pc.setRemoteDescription(
          new RTCSessionDescription(JSON.parse(atob(responseJson.sdp)))
        );
      } catch (e) {
        console.error(e);
      }
    }
  };

  // Offer to receive 1 video track and a data channel
  pc.addTransceiver("video", { direction: "recvonly" });

  const dataChannel = pc.createDataChannel("data", {
    ordered: false
  });
  dataChannel.onclose = () => console.log("Data channel closed");
  dataChannel.onopen = () => console.log("Data channel opened");
  dataChannel.onmessage = (msg) => {
    const event = BusEvent.decode(new Uint8Array(msg.data));
    busEventEmitter.emit(event);
    console.log(`emitted an bus event: ${event}`);
    console.log(event);
  };

  pc.createOffer()
    .then((d) => pc.setLocalDescription(d))
    .catch((e) => console.error(e));

  return [busEventEmitter, mediaStreamEmitter];
}
