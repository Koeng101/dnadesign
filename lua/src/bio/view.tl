
local genbank = require("genbank")
local record view
	sequence_view: function(genbank.Genbank, integer): (SequenceView, string)
end

local record SequenceView
	text: string
	features: {SequenceViewFeature}
end

local record SequenceViewFeature
	id: string
	type: string
	name: string
	spans: {SequenceViewFeatureSpan}
end

local record SequenceViewFeatureSpan
	line: integer
	start: integer
	end_exclusive: integer
end

function view.sequence_view(gb: genbank.Genbank, line_length: integer): (SequenceView, string)
	local _ = gb
	local _ = line_length
	return nil, nil
end

return view
