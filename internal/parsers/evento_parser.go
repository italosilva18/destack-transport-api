package parsers

import (
	"encoding/xml"
	"fmt"
)

// ParseEventoCTe faz o parse de um evento de CT-e
func ParseEventoCTe(xmlContent []byte) (*EventoParsed, error) {
	var procEvento ProcEventoCTe

	// Tentar fazer unmarshal do XML
	err := xml.Unmarshal(xmlContent, &procEvento)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do evento: %w", err)
	}

	// Extrair chave do CT-e
	chave := procEvento.EventoCTe.InfEvento.ChCTe
	if len(chave) != 44 {
		return nil, fmt.Errorf("chave de acesso inválida: %s", chave)
	}

	// Parsear data do evento
	dataEvento, err := ParseDate(procEvento.EventoCTe.InfEvento.DhEvento)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear data do evento: %w", err)
	}

	result := &EventoParsed{
		Chave:      chave,
		TipoEvento: procEvento.EventoCTe.InfEvento.TpEvento,
		Sequencia:  procEvento.EventoCTe.InfEvento.NSeqEvento,
		DataEvento: dataEvento,
		Protocolo:  procEvento.RetEventoCTe.InfEvento.NProt,
		Status:     procEvento.RetEventoCTe.InfEvento.CStat,
		Motivo:     procEvento.RetEventoCTe.InfEvento.XMotivo,
	}

	// Mapear tipo de evento
	tipoEventoMap := map[string]string{
		"110111": "Cancelamento",
		"110110": "Carta de Correção",
		"110140": "EPEC",
		"110170": "Cancelamento por Substituição",
	}

	if desc, ok := tipoEventoMap[result.TipoEvento]; ok {
		result.TipoEventoDesc = desc
	} else {
		result.TipoEventoDesc = "Evento " + result.TipoEvento
	}

	// Detalhes específicos do evento
	if procEvento.EventoCTe.InfEvento.DetEvento.EvCancCTe != nil {
		result.Justificativa = procEvento.EventoCTe.InfEvento.DetEvento.EvCancCTe.XJust
		result.ProtocoloRef = procEvento.EventoCTe.InfEvento.DetEvento.EvCancCTe.NProt
	}

	return result, nil
}
