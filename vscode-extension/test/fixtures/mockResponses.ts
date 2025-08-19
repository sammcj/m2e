/**
 * Mock API responses for testing M2E extension
 */

export interface MockConvertResponse {
    originalText: string;
    convertedText: string;
    metadata: {
        spellingChanges: number;
        unitChanges: number;
        processingTimeMs: number;
        fileType: string;
    };
}

/**
 * Generate mock response for conversion requests
 */
export function generateMockResponse(
    originalText: string, 
    fileType: string = 'text',
    convertUnits: boolean = true
): MockConvertResponse {
    let convertedText = originalText;
    let spellingChanges = 0;
    let unitChanges = 0;

    // Apply spelling conversions
    const spellingMap: { [key: string]: string } = {
        'color': 'colour',
        'colors': 'colours',
        'colored': 'coloured',
        'coloring': 'colouring',
        'colorful': 'colourful',
        'organize': 'organise',
        'organizes': 'organises',
        'organized': 'organised',
        'organizing': 'organising',
        'organization': 'organisation',
        'organizations': 'organisations',
        'center': 'centre',
        'centers': 'centres',
        'centered': 'centred',
        'centering': 'centring',
        'analyze': 'analyse',
        'analyzes': 'analyses',
        'analyzed': 'analysed',
        'analyzing': 'analysing',
        'analyzer': 'analyser',
        'realize': 'realise',
        'realizes': 'realises',
        'realized': 'realised',
        'realizing': 'realising',
        'realization': 'realisation',
        'aluminum': 'aluminium',
        'honor': 'honour',
        'honors': 'honours',
        'honored': 'honoured',
        'honoring': 'honouring',
        'flavor': 'flavour',
        'flavors': 'flavours',
        'flavored': 'flavoured',
        'flavoring': 'flavouring',
        'neighbor': 'neighbour',
        'neighbors': 'neighbours',
        'neighborhood': 'neighbourhood',
        'defense': 'defence',
        'offense': 'offence'
    };

    // Count and apply spelling changes
    for (const [american, british] of Object.entries(spellingMap)) {
        const regex = new RegExp(`\\b${american}\\b`, 'gi');
        const matches = convertedText.match(regex);
        if (matches) {
            spellingChanges += matches.length;
            convertedText = convertedText.replace(regex, (match) => {
                // Preserve case
                if (match === match.toUpperCase()) {
                    return british.toUpperCase();
                } else if (match === match.toLowerCase()) {
                    return british.toLowerCase();
                } else if (match[0] === match[0].toUpperCase()) {
                    return british.charAt(0).toUpperCase() + british.slice(1);
                }
                return british;
            });
        }
    }

    // Apply unit conversions if enabled
    if (convertUnits) {
        const unitConversions = [
            // Temperature
            { pattern: /(\d+(?:\.\d+)?)\s*°F\b/g, convert: (f: number) => `${Math.round((f - 32) * 5/9)}°C` },
            { pattern: /(\d+(?:\.\d+)?)\s*degrees?\s+fahrenheit/gi, convert: (f: number) => `${Math.round((f - 32) * 5/9)}°C` },
            
            // Distance
            { pattern: /(\d+(?:\.\d+)?)\s*feet?\b/g, convert: (ft: number) => `${(ft * 0.3048).toFixed(2)}m` },
            { pattern: /(\d+(?:\.\d+)?)\s*ft\b/g, convert: (ft: number) => `${(ft * 0.3048).toFixed(2)}m` },
            { pattern: /(\d+(?:\.\d+)?)\s*inches?\b/g, convert: (inch: number) => `${(inch * 2.54).toFixed(1)}cm` },
            { pattern: /(\d+(?:\.\d+)?)\s*in\b/g, convert: (inch: number) => `${(inch * 2.54).toFixed(1)}cm` },
            { pattern: /(\d+(?:\.\d+)?)\s*miles?\b/g, convert: (mile: number) => `${(mile * 1.609).toFixed(2)}km` },
            { pattern: /(\d+(?:\.\d+)?)\s*yards?\b/g, convert: (yard: number) => `${(yard * 0.914).toFixed(2)}m` },
            
            // Weight
            { pattern: /(\d+(?:\.\d+)?)\s*pounds?\b/g, convert: (lb: number) => `${(lb * 0.454).toFixed(2)}kg` },
            { pattern: /(\d+(?:\.\d+)?)\s*lbs?\b/g, convert: (lb: number) => `${(lb * 0.454).toFixed(2)}kg` },
            { pattern: /(\d+(?:\.\d+)?)\s*ounces?\b/g, convert: (oz: number) => `${(oz * 28.35).toFixed(1)}g` },
            { pattern: /(\d+(?:\.\d+)?)\s*oz\b/g, convert: (oz: number) => `${(oz * 28.35).toFixed(1)}g` },
            
            // Volume
            { pattern: /(\d+(?:\.\d+)?)\s*gallons?\b/g, convert: (gal: number) => `${(gal * 3.785).toFixed(2)}L` },
            { pattern: /(\d+(?:\.\d+)?)\s*gal\b/g, convert: (gal: number) => `${(gal * 3.785).toFixed(2)}L` },
            { pattern: /(\d+(?:\.\d+)?)\s*pints?\b/g, convert: (pint: number) => `${(pint * 0.473).toFixed(2)}L` },
            { pattern: /(\d+(?:\.\d+)?)\s*quarts?\b/g, convert: (qt: number) => `${(qt * 0.946).toFixed(2)}L` },
            
            // Pressure
            { pattern: /(\d+(?:\.\d+)?)\s*psi\b/g, convert: (psi: number) => `${(psi * 6.895).toFixed(1)}kPa` },
            
            // Area
            { pattern: /(\d+(?:\.\d+)?)\s*sq\s*ft\b/g, convert: (sqft: number) => `${(sqft * 0.093).toFixed(2)} sq m` },
            { pattern: /(\d+(?:\.\d+)?)\s*acres?\b/g, convert: (acre: number) => `${(acre * 0.405).toFixed(2)} hectares` }
        ];

        for (const conversion of unitConversions) {
            const matches = Array.from(convertedText.matchAll(conversion.pattern));
            if (matches.length > 0) {
                unitChanges += matches.length;
                convertedText = convertedText.replace(conversion.pattern, (match, value) => {
                    const numValue = parseFloat(value);
                    return conversion.convert(numValue);
                });
            }
        }
    }

    return {
        originalText,
        convertedText,
        metadata: {
            spellingChanges,
            unitChanges,
            processingTimeMs: Math.floor(Math.random() * 50) + 10, // 10-60ms
            fileType
        }
    };
}

/**
 * Predefined mock responses for specific test cases
 */
export const mockResponses = {
    health: {
        status: 'healthy',
        version: '1.0.0',
        timestamp: '2024-08-19T10:00:00Z'
    },

    simple: generateMockResponse('The color is great', 'text'),
    
    multiple: generateMockResponse('The color of the organization center was analyzed', 'text'),
    
    noChanges: generateMockResponse('The colour is great', 'text'),
    
    jsCode: generateMockResponse(`// The color is great
function analyzeColor(colorValue) {
    // Organize the data
    const center = findCenter(colorValue);
    return center;
}`, 'javascript'),

    jsCommentsOnly: {
        originalText: `// The color is great
function colorAnalyzer() {
    return "color";
}`,
        convertedText: `// The colour is great
function colorAnalyzer() {
    return "color";
}`,
        metadata: {
            spellingChanges: 1,
            unitChanges: 0,
            processingTimeMs: 15,
            fileType: 'javascript'
        }
    },

    withUnits: generateMockResponse('The room is 70°F and measures 10 feet by 12 feet', 'text'),
    
    largeText: generateMockResponse('The color organization '.repeat(100), 'text'),
    
    empty: generateMockResponse('', 'text'),
    
    unicode: generateMockResponse('The cölor is great', 'text'),

    // Error responses
    errors: {
        notFound: {
            error: 'Endpoint not found',
            status: 404
        },
        
        badRequest: {
            error: 'Invalid request format',
            status: 400
        },
        
        serverError: {
            error: 'Internal server error',
            status: 500
        },
        
        timeout: {
            error: 'Request timeout',
            status: 408
        }
    }
};

/**
 * Generate mock response for comments-only conversion
 */
export function generateCommentsOnlyResponse(
    originalText: string,
    fileType: string
): MockConvertResponse {
    let convertedText = originalText;
    let spellingChanges = 0;

    const lines = originalText.split('\n');
    const convertedLines = lines.map(line => {
        const trimmed = line.trim();
        
        // Check if line is a comment based on file type
        const isComment = 
            (fileType === 'javascript' || fileType === 'typescript') && (trimmed.startsWith('//') || trimmed.startsWith('/*') || trimmed.includes('*/')) ||
            (fileType === 'python') && trimmed.startsWith('#') ||
            (fileType === 'go') && trimmed.startsWith('//') ||
            (fileType === 'java') && (trimmed.startsWith('//') || trimmed.startsWith('/*')) ||
            (fileType === 'c' || fileType === 'cpp') && (trimmed.startsWith('//') || trimmed.startsWith('/*'));

        if (isComment) {
            // Apply spelling conversions only to comments
            let convertedLine = line;
            const spellingMap: { [key: string]: string } = {
                'color': 'colour',
                'organize': 'organise',
                'analyze': 'analyse',
                'center': 'centre'
            };

            for (const [american, british] of Object.entries(spellingMap)) {
                const regex = new RegExp(`\\b${american}\\b`, 'gi');
                const matches = convertedLine.match(regex);
                if (matches) {
                    spellingChanges += matches.length;
                    convertedLine = convertedLine.replace(regex, british);
                }
            }
            
            return convertedLine;
        }
        
        return line;
    });

    return {
        originalText,
        convertedText: convertedLines.join('\n'),
        metadata: {
            spellingChanges,
            unitChanges: 0,
            processingTimeMs: Math.floor(Math.random() * 30) + 10,
            fileType
        }
    };
}

/**
 * Response delays for different scenarios (in milliseconds)
 */
export const responseDelays = {
    fast: 10,
    normal: 50,
    slow: 200,
    timeout: 5000
};

/**
 * Mock server states for testing different scenarios
 */
export const serverStates = {
    healthy: { status: 'healthy', responsive: true },
    slow: { status: 'healthy', responsive: true, delay: responseDelays.slow },
    unresponsive: { status: 'error', responsive: false },
    starting: { status: 'starting', responsive: false },
    stopping: { status: 'stopping', responsive: false }
};